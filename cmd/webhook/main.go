// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/intel/network-resources-injector/pkg/webhook"
)

func main() {
	/* load configuration */
	port := flag.Int("port", 443, "The port on which to serve.")
	address := flag.String("bind-address", "0.0.0.0", "The IP address on which to listen for the --port port.")
	cert := flag.String("tls-cert-file", "cert.pem", "File containing the default x509 Certificate for HTTPS.")
	key := flag.String("tls-private-key-file", "key.pem", "File containing the default x509 private key matching --tls-cert-file.")
	flag.Parse()

	glog.Infof("starting mutating admission controller for network resources injection")

	/* init API client */
	webhook.SetupInClusterClient()

	/* register handlers */
	var httpServer *http.Server
	http.HandleFunc("/mutate", webhook.MutateHandler)

	/* start webhook server */
	go func() {
		for {
			httpServer = &http.Server{
				Addr: fmt.Sprintf("%s:%d", *address, *port),
			}
			err := httpServer.ListenAndServeTLS(*cert, *key)
			if err != nil {
				if err == http.ErrServerClosed {
					glog.Info("restarting server")
					continue
				} else {
					glog.Fatalf("error starting web server: %s", err.Error())
					break
				}
			}
		}
	}()

	/* watch the cert file and restart http sever if the file updated. */
	oldHashVal := ""
	for {
		hasher := sha512.New()
		s, err := ioutil.ReadFile(*cert)
		hasher.Write(s)
		if err != nil {
			glog.Fatalf("failed to read file %s: %s", *cert, err)
		}
		newHashVal := hex.EncodeToString(hasher.Sum(nil))
		if oldHashVal != "" && newHashVal != oldHashVal {
			glog.Info("get cert file update")
			if err := httpServer.Shutdown(context.Background()); err != nil {
				glog.Fatalf("http server shutdown: %v", err)
			}
		}
		oldHashVal = newHashVal
		time.Sleep(1 * time.Second)
	}
}
