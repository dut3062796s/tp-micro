package configer

import (
	"context"
	"encoding/json"

	tp "github.com/henrylee2cn/teleport"
	micro "github.com/xiaoenai/tp-micro"
	"github.com/xiaoenai/tp-micro/model/etcd"
)

var mgr = struct {
	etcdClient *etcd.Client
}{}

// InitMgr initializes a config manager.
func InitMgr(etcdClient *etcd.Client) {
	mgr.etcdClient = etcdClient
}

// PullCtrl returns a new PULL controller.
func PullCtrl() interface{} {
	return new(cfg)
}

type cfg struct {
	tp.PullCtx
}

var (
	rerrEtcdError = micro.RerrInternalServerError.Copy().SetMessage("Etcd Error")
	rerrNotFound  = micro.RerrNotFound.Copy().SetReason("Config is not exist")
)

func (c *cfg) List(*struct{}) ([]string, *tp.Rerror) {
	resp, err := mgr.etcdClient.Get(context.TODO(), KEY_PREFIX, etcd.WithPrefix())
	if err != nil {
		return nil, rerrEtcdError.Copy().SetReason(err.Error())
	}
	var r = make([]string, len(resp.Kvs))
	for i, kv := range resp.Kvs {
		r[i] = string(kv.Key)
	}
	return r, nil
}

func (c *cfg) Get(*struct{}) (string, *tp.Rerror) {
	key := c.Query().Get("config-key")
	resp, err := mgr.etcdClient.Get(context.TODO(), key)
	if err != nil {
		return "", rerrEtcdError.Copy().SetReason(err.Error())
	}
	if len(resp.Kvs) == 0 {
		return "", rerrNotFound
	}
	n := new(Node)
	json.Unmarshal(resp.Kvs[0].Value, n)
	if n.Config == "" {
		return "{}", nil
	}
	return n.Config, nil
}

// ConfigKV config key-value data.
type ConfigKV struct {
	Key   string
	Value string
}

func (c *cfg) Update(cfgKv *ConfigKV) (*struct{}, *tp.Rerror) {
	_, err := mgr.etcdClient.Put(context.TODO(), cfgKv.Key, (&Node{
		Initialized: true,
		Config:      cfgKv.Value,
	}).String())
	if err != nil {
		return nil, rerrEtcdError.Copy().SetReason(err.Error())
	}
	return nil, nil
}
