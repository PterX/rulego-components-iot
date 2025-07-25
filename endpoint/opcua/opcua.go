/*
 * Copyright 2024 The RuleGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package opcua

import (
	"context"
	"encoding/json"
	"log"
	"net/textproto"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/errors"
	"github.com/robfig/cron/v3"

	opcuaClient "github.com/rulego/rulego-components-iot/pkg/opcua_client"
	"github.com/rulego/rulego/api/types"
	endpointApi "github.com/rulego/rulego/api/types/endpoint"
	"github.com/rulego/rulego/components/base"
	"github.com/rulego/rulego/endpoint"
	"github.com/rulego/rulego/endpoint/impl"
	"github.com/rulego/rulego/utils/maps"
)

const Type = types.EndpointTypePrefix + "opcua"
const OPC_UA_DATA_MSG_TYPE = "OPC_UA_DATA"

// Endpoint 别名
type Endpoint = OpcUa

var _ endpointApi.Endpoint = (*Endpoint)(nil)

// 注册组件
func init() {
	_ = endpoint.Registry.Register(&Endpoint{})
}

type RequestMessage struct {
	headers    textproto.MIMEHeader
	body       []byte
	data       []opcuaClient.Data
	msg        *types.RuleMsg
	statusCode int
	err        error
}

func (r *RequestMessage) Body() []byte {
	b, err := json.Marshal(r.data)
	if err != nil {
		log.Println(err)
	}
	return b
}

func (r *RequestMessage) Headers() textproto.MIMEHeader {
	if r.headers == nil {
		r.headers = make(map[string][]string)
	}
	return r.headers
}

func (r *RequestMessage) From() string {
	return ""
}

// GetParam 不提供获取参数
func (r *RequestMessage) GetParam(key string) string {
	return ""
}

func (r *RequestMessage) SetMsg(msg *types.RuleMsg) {
	r.msg = msg
}
func (r *RequestMessage) GetMsg() *types.RuleMsg {
	if r.msg == nil {
		//默认指定是JSON格式，如果不是该类型，请在process函数中修改
		ruleMsg := types.NewMsg(0, OPC_UA_DATA_MSG_TYPE, types.JSON, types.NewMetadata(), string(r.Body()))
		r.msg = &ruleMsg
	}
	return r.msg
}

func (r *RequestMessage) SetStatusCode(statusCode int) {
	r.statusCode = statusCode
}
func (r *RequestMessage) SetBody(body []byte) {
	r.body = body
}

// SetError set error
func (r *RequestMessage) SetError(err error) {

}

// GetError get error
func (r *RequestMessage) GetError() error {
	return r.err
}

type ResponseMessage struct {
	headers    textproto.MIMEHeader
	body       []byte
	data       []opcuaClient.Data
	msg        *types.RuleMsg
	statusCode int
	err        error
}

func (r *ResponseMessage) Body() []byte {
	b, err := json.Marshal(r.body)
	if err != nil {
		log.Println(err)
	}
	return b
}

func (r *ResponseMessage) Headers() textproto.MIMEHeader {
	if r.headers == nil {
		r.headers = make(map[string][]string)
	}
	return r.headers
}

func (r *ResponseMessage) From() string {
	return ""
}

// GetParam 不提供获取参数
func (r *ResponseMessage) GetParam(key string) string {
	return ""
}

func (r *ResponseMessage) SetMsg(msg *types.RuleMsg) {
	r.msg = msg
}
func (r *ResponseMessage) GetMsg() *types.RuleMsg {
	if r.msg == nil {
		//默认指定是JSON格式，如果不是该类型，请在process函数中修改
		ruleMsg := types.NewMsg(0, OPC_UA_DATA_MSG_TYPE, types.JSON, types.NewMetadata(), string(r.Body()))
		r.msg = &ruleMsg
	}
	return r.msg
}

func (r *ResponseMessage) SetStatusCode(statusCode int) {
	r.statusCode = statusCode
}
func (r *ResponseMessage) SetBody(body []byte) {
	r.body = body
}
func (r *ResponseMessage) getBody() []byte {
	return r.body
}

// SetError set error
func (r *ResponseMessage) SetError(err error) {

}

// GetError get error
func (r *ResponseMessage) GetError() error {
	return r.err
}

// OpcUaConfig OPC UA Server配置
type OpcUaConfig struct {
	//OPC UA Server Endpoint, eg. opc.tcp://localhost:4840
	Server string `json:"server"`
	//Security Policy URL or one of None, Basic128Rsa15, Basic256, Basic256Sha256
	Policy string `json:"policy"`
	//Security Mode: one of None, Sign, SignAndEncrypt
	Mode string `json:"mode"`
	//Authentication Mode: one of Anonymous, UserName, Certificate
	Auth string `json:"auth"`
	//Authentication Username
	Username string `json:"username"`
	//Authentication Password
	Password string `json:"password"`
	//OPC UA Server CertFile Path
	CertFile string `json:"certFile"`
	//OPC UA Server CertKeyFile Path
	CertKeyFile string `json:"certKeyFile"`
	//Interval to read, supports cron expressions
	//example: @every 1m (every 1 minute) 0 0 0 * * * (triggers at midnight)
	Interval string `json:"interval"`
	//NodeIds to read, eg. ns=2;s=Channel1.Device1.Tag1
	NodeIds []string `json:"nodeIds"`
}

func (c OpcUaConfig) GetServer() string {
	return c.Server
}
func (c OpcUaConfig) GetPolicy() string {
	return c.Policy
}
func (c OpcUaConfig) GetMode() string {
	return c.Mode
}
func (c OpcUaConfig) GetAuth() string {
	return c.Auth
}
func (c OpcUaConfig) GetUsername() string {
	return c.Username
}
func (c OpcUaConfig) GetPassword() string {
	return c.Password
}
func (c OpcUaConfig) GetCertFile() string {
	return c.CertFile
}
func (c OpcUaConfig) GetCertKeyFile() string {
	return c.CertKeyFile
}

type OpcUa struct {
	impl.BaseEndpoint
	base.SharedNode[*opcua.Client]
	// GracefulShutdown provides graceful shutdown capabilities
	// GracefulShutdown 提供优雅停机功能
	base.GracefulShutdown
	RuleConfig types.Config
	// opcua client相关配置
	Config OpcUaConfig
	// 路由实例
	Router endpointApi.Router
	// 定时任务实例
	cronTask *cron.Cron
	// 定时任务id
	taskId cron.EntryID
}

// Type 组件类型
func (x *OpcUa) Type() string {
	return Type
}

// New 创建组件实例
func (x *OpcUa) New() types.Node {
	return &OpcUa{
		Config: OpcUaConfig{
			Interval: "@every 1m",
			Server:   "opc.tcp://localhost:4840",
			Policy:   "None",
			Mode:     "none",
			Auth:     "anonymous",
		},
	}
}

// Init 初始化
func (x *OpcUa) Init(ruleConfig types.Config, configuration types.Configuration) error {
	err := maps.Map2Struct(configuration, &x.Config)
	x.RuleConfig = ruleConfig

	// 初始化优雅停机功能 - 使用合理的默认超时(10秒)
	x.GracefulShutdown.InitGracefulShutdown(x.RuleConfig.Logger, 10*time.Second)

	_ = x.SharedNode.InitWithClose(x.RuleConfig, x.Type(), x.Config.Server, true, func() (*opcua.Client, error) {
		return x.initClient()
	}, func(client *opcua.Client) error {
		if client != nil {
			return client.Close(context.Background())
		}
		return nil
	})
	return err
}

// Destroy 销毁
func (x *OpcUa) Destroy() {
	x.GracefulShutdown.GracefulStop(func() {
		_ = x.Close()
	})
}

// GracefulStop provides graceful shutdown for the OPC UA endpoint
// GracefulStop 为 OPC UA 端点提供优雅停机
func (x *OpcUa) GracefulStop() {
	x.GracefulShutdown.GracefulStop(func() {
		_ = x.Close()
	})
}

func (x *OpcUa) Close() error {
	if x.taskId != 0 && x.cronTask != nil {
		x.cronTask.Remove(x.taskId)
	}
	if x.cronTask != nil {
		x.cronTask.Stop()
	}
	// SharedNode 会通过 InitWithClose 中的清理函数来管理客户端的关闭
	// SharedNode manages client closure through the cleanup function in InitWithClose
	_ = x.SharedNode.Close()
	return nil
}

func (x *OpcUa) Id() string {
	return x.Config.Server
}

func (x *OpcUa) AddRouter(router endpointApi.Router, params ...interface{}) (string, error) {
	if router == nil {
		return "", errors.New("router cannot be nil")
	}
	if x.Router != nil {
		return "", errors.New("duplicate router")
	}
	x.Router = router
	return router.GetId(), nil
}

func (x *OpcUa) RemoveRouter(routerId string, params ...interface{}) error {
	x.Lock()
	defer x.Unlock()
	x.Router = nil
	return nil
}

func (x *OpcUa) Start() error {
	var err error
	if !x.SharedNode.IsInit() {
		err = x.SharedNode.InitWithClose(x.RuleConfig, x.Type(), x.Config.Server, true, func() (*opcua.Client, error) {
			return x.initClient()
		}, func(client *opcua.Client) error {
			if client != nil {
				return client.Close(context.Background())
			}
			return nil
		})
	}
	if x.cronTask != nil {
		x.cronTask.Stop()
	}
	x.cronTask = cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger)), cron.WithLogger(cron.DefaultLogger))
	eid, err := x.cronTask.AddFunc(x.Config.Interval, func() {
		if x.Router != nil {
			_ = x.readNodes(x.Router)
		}
	})
	x.taskId = eid
	x.cronTask.Start()
	return err
}

func (x *OpcUa) Printf(format string, v ...interface{}) {
	if x.RuleConfig.Logger != nil {
		x.RuleConfig.Logger.Printf(format, v...)
	}
}

func (x *OpcUa) readNodes(router endpointApi.Router) error {
	// 增加活跃操作计数
	x.GracefulShutdown.IncrementActiveOperations()
	defer x.GracefulShutdown.DecrementActiveOperations()

	client, err := x.SharedNode.GetSafely()
	if err != nil {
		x.Printf("get shared client error %v ", err)
		return err
	}

	data, _, err := opcuaClient.Read(client, x.Config.NodeIds)
	if err != nil {
		x.Printf("read nodes error %v ", err)
		return err
	}
	exchange := &endpointApi.Exchange{
		In: &RequestMessage{data: data},
		Out: &ResponseMessage{
			data: data,
		}}

	x.DoProcess(context.Background(), router, exchange)
	return nil
}

// initClient 初始化客户端
func (x *OpcUa) initClient() (*opcua.Client, error) {
	return opcuaClient.DefaultHolder(x.Config).NewOpcUaClient()
}
