package memsto

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ccfos/nightingale/v6/alert/aconf"
	"github.com/ccfos/nightingale/v6/dumper"
	"github.com/ccfos/nightingale/v6/models"
	"github.com/ccfos/nightingale/v6/pkg/ctx"
	"github.com/ccfos/nightingale/v6/pkg/poster"
	"github.com/ccfos/nightingale/v6/pkg/tplx"

	"github.com/BurntSushi/toml"
	"github.com/toolkits/pkg/logger"
)

type NotifyConfigCacheType struct {
	ctx         *ctx.Context
	ConfigCache *ConfigCache
	webhooks    map[string]*models.Webhook
	smtp        aconf.SMTPConfig
	script      models.NotifyScript

	sync.RWMutex
}

const DefaultSMTP = `
Host = ""
Port = 994
User = "username"
Pass = "password"
From = "username@163.com"
InsecureSkipVerify = true
Batch = 5
`

const DefaultIbex = `
Address = "http://127.0.0.1:10090"
BasicAuthUser = "ibex"
BasicAuthPass = "ibex"
Timeout = 3000
`

func NewNotifyConfigCache(ctx *ctx.Context, configCache *ConfigCache) *NotifyConfigCacheType {
	w := &NotifyConfigCacheType{
		ctx:         ctx,
		ConfigCache: configCache,
		webhooks:    make(map[string]*models.Webhook),
	}
	w.SyncNotifyConfigs()
	return w
}

func (w *NotifyConfigCacheType) SyncNotifyConfigs() {
	err := w.syncNotifyConfigs()
	if err != nil {
		logger.Error("failed to sync webhooks:", err)
	}

	go w.loopSyncNotifyConfigs()
}

func (w *NotifyConfigCacheType) loopSyncNotifyConfigs() {
	duration := time.Duration(9000) * time.Millisecond
	for {
		time.Sleep(duration)
		if err := w.syncNotifyConfigs(); err != nil {
			logger.Warning("failed to sync webhooks:", err)
		}
	}
}

func (w *NotifyConfigCacheType) syncNotifyConfigs() error {
	start := time.Now()
	userVariableMap := w.ConfigCache.Get()

	w.RWMutex.Lock()
	defer w.RWMutex.Unlock()

	cval, err := models.ConfigsGet(w.ctx, models.WEBHOOKKEY)
	if err != nil {
		dumper.PutSyncRecord("webhooks", start.Unix(), -1, -1, "failed to query configs.webhook: "+err.Error())
		return err
	}

	if strings.TrimSpace(cval) != "" {
		var webhooks []*models.Webhook
		err = json.Unmarshal([]byte(cval), &webhooks)
		if err != nil {
			dumper.PutSyncRecord("webhooks", start.Unix(), -1, -1, "failed to unmarshal configs.webhook: "+err.Error())
			logger.Errorf("failed to unmarshal webhooks:%s error:%v", cval, err)
		}

		newWebhooks := make(map[string]*models.Webhook, len(webhooks))
		for i := 0; i < len(webhooks); i++ {
			if webhooks[i].Batch == 0 {
				webhooks[i].Batch = 1000
			}

			if webhooks[i].Timeout == 0 {
				webhooks[i].Timeout = 10
			}

			if webhooks[i].RetryCount == 0 {
				webhooks[i].RetryCount = 10
			}

			if webhooks[i].RetryInterval == 0 {
				webhooks[i].RetryInterval = 10
			}

			if webhooks[i].Client == nil {
				transport := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: webhooks[i].SkipVerify},
				}
				if poster.UseProxy(webhooks[i].Url) {
					transport.Proxy = http.ProxyFromEnvironment
				}
				webhooks[i].Client = &http.Client{
					Timeout:   time.Second * time.Duration(webhooks[i].Timeout),
					Transport: transport,
				}
			}

			newWebhooks[webhooks[i].Url] = webhooks[i]
		}

		for url, wh := range newWebhooks {
			if oldWh, has := w.webhooks[url]; has && oldWh.Hash() != wh.Hash() {
				w.webhooks[url] = wh
			} else {
				w.webhooks[url] = wh
			}
		}

		for url := range w.webhooks {
			if _, has := newWebhooks[url]; !has {
				delete(w.webhooks, url)
			}
		}
	}

	dumper.PutSyncRecord("webhooks", start.Unix(), time.Since(start).Milliseconds(), len(w.webhooks), "success, webhooks:\n"+cval)

	start = time.Now()
	cval, err = models.ConfigsGet(w.ctx, models.SMTP)
	if err != nil {
		dumper.PutSyncRecord("smtp", start.Unix(), -1, -1, "failed to query configs.smtp_config: "+err.Error())
		return err
	}

	cval = tplx.ReplaceTemplateUseText(models.SMTP, cval, userVariableMap)

	if strings.TrimSpace(cval) != "" {
		err = toml.Unmarshal([]byte(cval), &w.smtp)
		if err != nil {
			dumper.PutSyncRecord("smtp", start.Unix(), -1, -1, "failed to unmarshal configs.smtp_config: "+err.Error())
			logger.Errorf("failed to unmarshal smtp:%s error:%v", cval, err)
		}
	}

	dumper.PutSyncRecord("smtp", start.Unix(), time.Since(start).Milliseconds(), 1, "success, smtp_config:\n"+cval)

	start = time.Now()
	cval, err = models.ConfigsGet(w.ctx, models.NOTIFYSCRIPT)
	if err != nil {
		dumper.PutSyncRecord("notify_script", start.Unix(), -1, -1, "failed to query configs.notify_script: "+err.Error())
		return err
	}

	if strings.TrimSpace(cval) != "" {
		err = json.Unmarshal([]byte(cval), &w.script)
		if err != nil {
			dumper.PutSyncRecord("notify_script", start.Unix(), -1, -1, "failed to unmarshal configs.notify_script: "+err.Error())
			logger.Errorf("failed to unmarshal notify script:%s error:%v", cval, err)
		}
	}

	dumper.PutSyncRecord("notify_script", start.Unix(), time.Since(start).Milliseconds(), 1, "success, notify_script:\n"+cval)

	return nil
}

func (w *NotifyConfigCacheType) GetWebhooks() map[string]*models.Webhook {
	w.RWMutex.RLock()
	defer w.RWMutex.RUnlock()
	return w.webhooks
}

func (w *NotifyConfigCacheType) GetSMTP() aconf.SMTPConfig {
	w.RWMutex.RLock()
	defer w.RWMutex.RUnlock()
	return w.smtp
}

func (w *NotifyConfigCacheType) GetNotifyScript() models.NotifyScript {
	w.RWMutex.RLock()
	defer w.RWMutex.RUnlock()
	if w.script.Timeout == 0 {
		w.script.Timeout = 10
	}

	return w.script
}
