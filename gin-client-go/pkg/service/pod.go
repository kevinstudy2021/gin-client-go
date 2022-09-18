package service

import (
	"awesomeProject1/gin-client-go/pkg/client"
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	"net/http"
	"sync"
)

func GetPod(ns string) ([]v1.Pod, error) {
	ctx := context.Background()
	clientSet, err := client.GetK8SClientSet()
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	podList, err := clientSet.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return podList.Items, nil
}

type WsMessage struct {
	MessageType int
	Data        []byte
}

type WsConnection struct {
	wsSocket  *websocket.Conn
	inChan    chan *WsMessage
	outChan   chan *WsMessage
	mutex     sync.Mutex
	isClosed  bool
	closeChan chan byte
}

func (wsConn *WsConnection) Wsclose() {
	err := wsConn.wsSocket.Close()
	if err != nil {
		klog.Errorln(err)
		return
	}
	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()
	if !wsConn.isClosed {
		wsConn.isClosed = true
		close(wsConn.closeChan)
	}
}

func (wsConn *WsConnection) wsReadLoop() {
	var (
		msgType int
		data    []byte
		msg     *WsMessage
		err     error
	)
	for {
		if msgType, data, err = wsConn.wsSocket.ReadMessage(); err != nil {
			goto ERROR
		}
		msg = &WsMessage{
			MessageType: msgType,
			Data:        data,
		}
		select {
		case wsConn.inChan <- msg:
		case <-wsConn.closeChan:
			goto CLOSED
		}
	}
ERROR:
	wsConn.Wsclose()
CLOSED:
}

func (wsConn *WsConnection) wsWriteLoop() {
	var (
		msg *WsMessage
		err error
	)
	for {
		select {
		case msg = <-wsConn.outChan:
			if err = wsConn.wsSocket.WriteMessage(msg.MessageType, msg.Data); err != nil {
				goto ERROR
			}
		case <-wsConn.closeChan:
			goto CLOSED
		}
	}
ERROR:
	wsConn.Wsclose()
CLOSED:
}

func (wsConn *WsConnection) WsWrite(messageType int, data []byte) (err error) {
	select {
	case wsConn.outChan <- &WsMessage{MessageType: messageType, Data: data}:
		return
	case <-wsConn.closeChan:
		err = errors.New("websocket closed")
	}
	return
}

func (wsConn *WsConnection) WsRead() (msg *WsMessage, err error) {
	select {
	case msg = <-wsConn.inChan:
		return
	case <-wsConn.closeChan:
		err = errors.New("websocket closed")
	}
	return
}

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func InitWebsocket(resp http.ResponseWriter, req *http.Request) (wsConn *WsConnection, err error) {
	var (
		wsSocket *websocket.Conn
	)
	if wsSocket, err = wsUpgrader.Upgrade(resp, req, nil); err != nil {
		klog.Errorln(err)
		return
	}
	wsConn = &WsConnection{
		wsSocket:  wsSocket,
		inChan:    make(chan *WsMessage, 1000),
		outChan:   make(chan *WsMessage, 1000),
		closeChan: make(chan byte),
		isClosed:  false,
	}
	// du qu xiecheng
	go wsConn.wsReadLoop()
	// xie xiecheng
	go wsConn.wsWriteLoop()
	return
}

type streamHandler struct {
	WsConn *WsConnection
	resizeEvent chan remotecommand.TerminalSize
}

func (handler *streamHandler) Write(p []byte) (size int, err error) {
	copyData := make([]byte, len(p))
	copy(copyData, p)
	size = len(p)
	err = handler.WsConn.WsWrite(websocket.TextMessage, copyData)
	return
}

type xtermMessage struct {
	MsgType string `json:"type"`
	Input string `json:"input"`
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

func (handler *streamHandler) Read(p []byte) (size int, err error){
	var (
		xtermMsg xtermMessage
		msg *WsMessage
	)

	if msg, err = handler.WsConn.WsRead();err != nil{
		klog.Errorln(err)
		return
	}
	// jie xi
	if err = json.Unmarshal(msg.Data, &xtermMsg); err != nil{
		return
	}
	if xtermMsg.MsgType == "resize" {
		handler.resizeEvent <- remotecommand.TerminalSize{Width: xtermMsg.Cols, Height: xtermMsg.Rows}
	}else if xtermMsg.MsgType == "input" {
		size = len(xtermMsg.Input)
		copy(p, xtermMsg.Input)
	}
	return
}

func (handler *streamHandler)Next() (size *remotecommand.TerminalSize) {
	ret := <- handler.resizeEvent
	size = &ret
	return
}

func WebSSH(namespaceName,podName,containerName,method string, resp http.ResponseWriter,req *http.Request) error {
	var (
		err error
		executor remotecommand.Executor
		handler *streamHandler
		wsConn *WsConnection
	)
	config, err := client.GetRestConfig()
	if err != nil {
		return err
	}
	// xiu gai dian
	//clientSet := client.KubeClientSet
	clientSet, err := client.GetK8SClientSet()
	if err != nil {
		klog.Errorln(err)
		return err
	}
	reqSSH := clientSet.CoreV1().RESTClient().Post().Resource("pods").Name(podName).Namespace(namespaceName).SubResource("exec").VersionedParams(
		&v1.PodExecOptions{
			Container: containerName,
			Command: []string{method},
			Stderr: true,
			Stdout: true,
			Stdin: true,
			TTY: true,
		},scheme.ParameterCodec)
	if executor, err = remotecommand.NewSPDYExecutor(config, "POST", reqSSH.URL()); err !=nil{
		klog.Errorln(err)
		return err
	}
	if wsConn, err = InitWebsocket(resp, req);err != nil{
		return err
	}
	handler = &streamHandler{WsConn: wsConn, resizeEvent: make(chan remotecommand.TerminalSize)}
	if err = executor.Stream(remotecommand.StreamOptions{
		Stdin: handler,
		Stdout: handler,
		Stderr: handler,
		TerminalSizeQueue: handler,
		Tty: true,
	});err != nil{
		goto END
	}

END:
	klog.Errorln(err)
	wsConn.Wsclose()
	return err

}
