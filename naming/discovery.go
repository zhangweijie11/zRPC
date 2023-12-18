package naming

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	_registerURL = "http://%s/api/register"
	_cancelURL   = "http://%s/api/cancel"
	_renewURL    = "http://%s/api/renew"
	_fetchURL    = "http://%s/api/fetch"
	_nodesURL    = "http://%s/api/nodes"
)

const (
	NodeInterval  = 90 * time.Second
	RenewInterval = 60 * time.Second
)

type Config struct {
	Nodes []string
	Env   string
}

type FetchData struct {
	Instances       []*Instance `json:"instances"`
	LatestTimestamp int64       `json:"latest_timestamp"`
}

type Discovery struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	once       *sync.Once
	conf       *Config
	// 本地缓存
	mutex    sync.RWMutex
	apps     map[string]*FetchData
	registry map[string]struct{}
	// 注册中心
	idx  uint64       // 节点索引
	node atomic.Value // 节点列表
}

type Response struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type ResponseFetch struct {
	Response
	Data FetchData `json:"data"`
}

// NewDiscovery 初始化服务注册发现
func NewDiscovery(conf *Config) *Discovery {
	if len(conf.Nodes) == 0 {
		panic("节点配置为空")
	}
	ctx, cancel := context.WithCancel(context.Background())
	discovery := &Discovery{
		ctx:        ctx,
		cancelFunc: cancel,
		once:       nil,
		conf:       conf,
		mutex:      sync.RWMutex{},
		apps:       map[string]*FetchData{},
		registry:   map[string]struct{}{},
		idx:        0,
		node:       atomic.Value{},
	}

	discovery.node.Store(conf.Nodes)

	go discovery.updateNode()

	return discovery
}

// 选择节点
func (dis *Discovery) pickNode() string {
	nodes, ok := dis.node.Load().([]string)
	if !ok || len(nodes) == 0 {
		// 返回默认节点
		return dis.conf.Nodes[dis.idx%uint64(len(dis.conf.Nodes))]
	}
	// 成功加载节点且不为空
	return nodes[dis.idx%uint64(len(nodes))]
}

// HttpPost 通用请求
func HttpPost(url string, data interface{}) (string, error) {
	client := &http.Client{}
	js, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", url, bytes.NewReader(js))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(body), nil
}

// 对比两个数据是否相等
func compareNodes(old, new []string) bool {
	if len(old) != len(new) {
		return false
	}
	mapB := make(map[string]struct{}, len(new))
	for _, node := range new {
		mapB[node] = struct{}{}
	}
	for _, node := range old {
		if _, ok := mapB[node]; !ok {
			return false
		}
	}
	return true
}

// 默认从配置中获取注册中心节点，并开启单独协程来定期更新维护节点变化
func (dis *Discovery) updateNode() {
	ticker := time.NewTicker(NodeInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			url := fmt.Sprintf(_nodesURL, dis.pickNode())
			log.Println("服务注册发现 - 请求并更新节点, 地址:" + url)
			params := make(map[string]interface{})
			params["env"] = dis.conf.Env
			response, err := HttpPost(url, params)
			if err != nil {
				log.Println(err)
				continue
			}

			responseFetch := ResponseFetch{}
			err = json.Unmarshal([]byte(response), &responseFetch)
			if err != nil {
				log.Println(err)
				continue
			}

			var newNodes []string
			for _, instance := range responseFetch.Data.Instances {
				for _, address := range instance.Addresses {
					newNodes = append(newNodes, strings.TrimPrefix(address, "http://"))
				}
			}
			if len(newNodes) == 0 {
				continue
			}

			curNodes := dis.node.Load().([]string)
			if !compareNodes(curNodes, newNodes) {
				// 存储新节点
				dis.node.Store(newNodes)
				log.Println("节点列表已更新！", newNodes)
				log.Println(newNodes)
			} else {
				log.Println("节点列表无更新！", curNodes)
			}
		}
	}
}

// 切换节点
func (dis *Discovery) switchNode() {
	atomic.AddUint64(&dis.idx, 1)
}

// 注册节点
func (dis *Discovery) register(instance *Instance) error {
	url := fmt.Sprintf(_registerURL, dis.pickNode())
	log.Println("服务注册发现 - 请求并更新节点, 地址:" + url)
	params := make(map[string]interface{})
	params["env"] = dis.conf.Env
	params["appid"] = instance.AppID
	params["hostname"] = instance.Hostname
	params["addresses"] = instance.Addresses
	params["version"] = instance.Version
	params["status"] = 1

	response, err := HttpPost(url, params)
	if err != nil {
		dis.switchNode()
		log.Println("注册出现异常: ", err, "地址: ", url)
		return err
	}

	resp := Response{}
	err = json.Unmarshal([]byte(response), &resp)
	if err != nil {
		return err
	}

	if resp.Code != 200 {
		log.Printf("注册出现异常, 地址为 (%v), 响应码为 (%v)\n", url, resp.Code)
		return errors.New("注册出现异常")
	}

	return err
}

// 重新注册
func (dis *Discovery) renew(instance *Instance) error {
	uri := fmt.Sprintf(_renewURL, dis.pickNode())
	log.Println("服务注册发现 - 重新注册地址:" + uri)
	params := make(map[string]interface{})
	params["env"] = dis.conf.Env
	params["appid"] = instance.AppID
	params["hostname"] = instance.Hostname

	response, err := HttpPost(uri, params)
	if err != nil {
		dis.switchNode()
		log.Println("重新注册出现异常: ", err, "地址: ", uri)
		return err
	}
	resp := Response{}
	err = json.Unmarshal([]byte(response), &resp)
	if err != nil {
		log.Println(err)
		return err
	}
	if resp.Code != 200 {
		log.Printf("重新注册地址为 (%v), 响应状态码为 (%v)\n", uri, resp.Code)
		return errors.New("not found")
	}
	return nil
}

// 关闭服务
func (dis *Discovery) cancel(instance *Instance) error {
	//local cache
	url := fmt.Sprintf(_cancelURL, dis.pickNode())
	log.Println("服务注册发现 - 请求关闭地址:" + url)
	params := make(map[string]interface{})
	params["env"] = dis.conf.Env
	params["appid"] = instance.AppID
	params["hostname"] = instance.Hostname

	response, err := HttpPost(url, params)
	if err != nil {
		dis.switchNode()
		log.Println("cancel err: ", err, "url: ", url)
		return err
	}
	resp := Response{}
	err = json.Unmarshal([]byte(response), &resp)
	if err != nil {
		log.Println(err)
		return err
	}
	if resp.Code != 200 {
		log.Printf("地址为 (%v), 响应状态码为 (%v)\n", url, resp.Code)
		return err
	}
	return nil
}

func (dis *Discovery) Register(ctx context.Context, instance *Instance) (context.CancelFunc, error) {
	var err error
	// 检查是否存在本地缓存
	dis.mutex.Lock()
	if _, ok := dis.registry[instance.AppID]; ok {
		err = errors.New("重复注册实例")
	} else {
		dis.registry[instance.AppID] = struct{}{}
	}
	dis.mutex.Unlock()

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(dis.ctx)
	if err = dis.register(instance); err != nil {
		dis.mutex.Lock()
		delete(dis.registry, instance.AppID)
		dis.mutex.Unlock()
		return cancel, err
	}

	ch := make(chan struct{}, 1)
	cancelFunc := context.CancelFunc(func() {
		cancel()
		<-ch
	})

	go func() {
		ticker := time.NewTicker(RenewInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err = dis.renew(instance); err != nil {
					dis.register(instance)
				}
			case <-ctx.Done():
				dis.cancel(instance)
				ch <- struct{}{}
			}
		}
	}()

	return cancelFunc, nil
}

// Fetch 根据服务标识获取服务注册信息，先从本地缓存获取，如果不存在再从远程注册中心获取并缓存
func (dis *Discovery) Fetch(ctx context.Context, appId string) ([]*Instance, bool) {
	dis.mutex.Lock()
	fetchData, ok := dis.apps[appId]
	dis.mutex.Unlock()

	if ok {
		log.Println("从本地缓存获取数据, appid:" + appId)
		return fetchData.Instances, ok
	}
	// 从远程注册中心获取
	uri := fmt.Sprintf(_fetchURL, dis.pickNode())
	params := make(map[string]interface{})
	params["env"] = dis.conf.Env
	params["appid"] = appId
	params["status"] = 1 //up
	resp, err := HttpPost(uri, params)
	if err != nil {
		dis.switchNode()
		return nil, false
	}
	res := ResponseFetch{}
	err = json.Unmarshal([]byte(resp), &res)
	if res.Code != 200 {
		return nil, false
	}
	if err != nil {
		log.Println(err)
		return nil, false
	}
	var result []*Instance
	for _, ins := range res.Data.Instances {
		result = append(result, ins)

	}
	if len(result) > 0 {
		ok = true
		dis.mutex.Lock()
		dis.apps[appId] = &res.Data
		dis.mutex.Unlock()
	}
	return result, ok
}
