package apollo_client

import (
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/tianyabin2010/ludo-gobase/util"
	apollo "github.com/zouyx/agollo/v4"
	"github.com/zouyx/agollo/v4/env/config"
	"github.com/zouyx/agollo/v4/storage"
)

var (
	defaultApolloListener = &apolloListener{}
	apolloClient          *apollo.Client
)

type configUpdateFunc func(string, []byte) error

type apolloListener struct {
	NameSpace string
	Callback  configUpdateFunc
}

func (l *apolloListener) OnChange(event *storage.ChangeEvent) {
	util.BtRecover("apolloListener.OnChange")
	if event.Namespace == l.NameSpace {
		if nil != l.Callback {
			for k, v := range event.Changes {
				data, ok := v.NewValue.(string)
				if ok {
					if err := l.Callback(k, []byte(data)); err == nil {
						log.Info().Str("key", k).Interface("val", v).
							Msgf("onchange config update success: %v", k)
					} else {
						log.Error().Err(err).Str("key", k).Interface("val", v).
							Msgf("onchange config update error: %v", k)
					}
				} else {
					log.Error().
						Str("key", k).
						Interface("val", v.NewValue).
						Msgf("apollo onchange newval type error")
				}
			}
		}
	}
}

func (l *apolloListener) OnNewestChange(event *storage.FullChangeEvent) {
	util.BtRecover("apolloListener.OnNewestChange")
	if event.Namespace == l.NameSpace {
		if nil != l.Callback {
			for k, v := range event.Changes {
				data, ok := v.(string)
				if ok {
					if err := l.Callback(k, []byte(data)); err == nil {
						log.Info().Str("key", k).Interface("val", v).
							Msgf("onnewwestchange config update success: %v", k)
					} else {
						log.Error().Err(err).Str("key", k).Interface("val", v).
							Msgf("onnewwestchange config update error: %v", k)
					}
				} else {
					log.Error().
						Str("key", k).
						Interface("val", v).
						Msgf("apollo on newest change newval type error")
				}
			}
		}
	}
}

func Init(ApolloAddr, AppId, Cluster, NameSpace string, ConfigUpdate configUpdateFunc) bool {
	c := &config.AppConfig{
		AppID:         AppId,
		Cluster:       Cluster,
		IP:            ApolloAddr,
		NamespaceName: NameSpace,
	}
	var err error
	apolloClient, err = apollo.StartWithConfig(func() (*config.AppConfig, error) {
		return c, nil
	})
	if err != nil {
		log.Error().Err(err).Msgf("config init error")
		panic(err)
	}
	defaultApolloListener.Callback = ConfigUpdate
	defaultApolloListener.NameSpace = NameSpace
	apolloClient.AddChangeListener(defaultApolloListener)
	cache := apolloClient.GetConfigCache(NameSpace)
	bInit := true
	if nil != cache {
		cache.Range(func(key, val interface{}) bool {
			k, ok := key.(string)
			if !ok {
				log.Error().Interface("key", key).Msgf("config init key type error")
				return true
			}
			v, ok := val.(string)
			if !ok {
				log.Error().Interface("val", val).Msgf("config init val type error")
				return true
			}
			if nil != ConfigUpdate {
				if err := ConfigUpdate(k, []byte(v)); err == nil {
					log.Info().Str("key", k).Str("val", string(v)).
						Msgf("config init success: %v", k)
				} else {
					log.Error().Err(err).Str("key", k).Str("val", string(v)).
						Msgf("config init error: %v", k)
					bInit = false
				}
			}
			return true
		})
	}
	return bInit
}

type ApolloClient struct {
	client *apollo.Client
	listener *apolloListener
}

func CreateApolloClient(ApolloAddr, AppId, Cluster, NameSpace string, ConfigUpdate configUpdateFunc) (*ApolloClient, error) {
	c := &config.AppConfig{
		AppID:         AppId,
		Cluster:       Cluster,
		IP:            ApolloAddr,
		NamespaceName: NameSpace,
	}
	cli, err := apollo.StartWithConfig(func() (*config.AppConfig, error) {
		return c, nil
	})
	if err != nil {
		log.Error().Err(err).Msgf("config init error")
		panic(err)
	}
	listener := &apolloListener{}
	listener.Callback = ConfigUpdate
	listener.NameSpace = NameSpace
	cli.AddChangeListener(listener)
	cache := cli.GetConfigCache(NameSpace)
	bInit := true
	if nil != cache {
		cache.Range(func(key, val interface{}) bool {
			k, ok := key.(string)
			if !ok {
				log.Error().Interface("key", key).Msgf("config init key type error")
				return true
			}
			v, ok := val.(string)
			if !ok {
				log.Error().Interface("val", val).Msgf("config init val type error")
				return true
			}
			if nil != ConfigUpdate {
				if err := ConfigUpdate(k, []byte(v)); err == nil {
					log.Info().Str("key", k).Str("val", string(v)).
						Msgf("config init success: %v", k)
				} else {
					log.Error().Err(err).Str("key", k).Str("val", string(v)).
						Msgf("config init error: %v", k)
					bInit = false
				}
			}
			return true
		})
	}
	if bInit {
		err = nil
	} else {
		err = errors.New("apollo client create error")
	}
	return &ApolloClient{
		client:   cli,
		listener: listener,
	}, err
}