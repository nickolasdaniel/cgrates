/*
Real-time Online/Offline Charging System (OCS) for Telecom & ISP environments
Copyright (C) ITsysCOM GmbH

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package services

import (
	"fmt"
	"sync"

	"github.com/cgrates/cgrates/agents"
	"github.com/cgrates/cgrates/servmanager"
	"github.com/cgrates/cgrates/sessions"
	"github.com/cgrates/cgrates/utils"
	"github.com/cgrates/rpcclient"
)

// NewFreeswitchAgent returns the Freeswitch Agent
func NewFreeswitchAgent() servmanager.Service {
	return new(FreeswitchAgent)
}

// FreeswitchAgent implements Agent interface
type FreeswitchAgent struct {
	sync.RWMutex
	fS *agents.FSsessions
}

// Start should handle the sercive start
func (fS *FreeswitchAgent) Start(sp servmanager.ServiceProvider, waitCache bool) (err error) {
	if fS.IsRunning() {
		return fmt.Errorf("service aleady running")
	}

	fS.Lock()
	defer fS.Unlock()
	var sS rpcclient.RpcClientConnection
	var sSInternal bool
	utils.Logger.Info("Starting FreeSWITCH agent")
	if !sp.GetConfig().DispatcherSCfg().Enabled && sp.GetConfig().FsAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		sSInternal = true
		srvSessionS, has := sp.GetService(utils.SessionS)
		if !has {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed to find needed subsystem <%s>",
				utils.FreeSWITCHAgent, utils.SessionS))
			return utils.ErrNotFound
		}
		sSIntConn := <-srvSessionS.GetIntenternalChan()
		srvSessionS.GetIntenternalChan() <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		if sS, err = sp.NewConnection(utils.SessionS, sp.GetConfig().FsAgentCfg().SessionSConns); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.FreeSWITCHAgent, utils.SessionS, err.Error()))
			return
		}
	}
	fS.fS = agents.NewFSsessions(sp.GetConfig().FsAgentCfg(), sS, sp.GetConfig().GeneralCfg().DefaultTimezone)
	if sSInternal { // bidirectional client backwards connection
		sS.(*utils.BiRPCInternalClient).SetClientConn(fS.fS)
		var rply string
		if err = sS.Call(utils.SessionSv1RegisterInternalBiJSONConn,
			utils.EmptyString, &rply); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.FreeSWITCHAgent, utils.SessionS, err.Error()))
			return
		}
	}

	go func() {
		if err := fS.fS.Connect(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.FreeSWITCHAgent, err))
			sp.GetExitChan() <- true // stop the engine here
		}
	}()
	return
}

// GetIntenternalChan returns the internal connection chanel
// no chanel for FreeswitchAgent
func (fS *FreeswitchAgent) GetIntenternalChan() (conn chan rpcclient.RpcClientConnection) {
	return nil
}

// Reload handles the change of config
func (fS *FreeswitchAgent) Reload(sp servmanager.ServiceProvider) (err error) {
	var sS rpcclient.RpcClientConnection
	var sSInternal bool
	if !sp.GetConfig().DispatcherSCfg().Enabled && sp.GetConfig().FsAgentCfg().SessionSConns[0].Address == utils.MetaInternal {
		sSInternal = true
		srvSessionS, has := sp.GetService(utils.SessionS)
		if !has {
			utils.Logger.Err(fmt.Sprintf("<%s> Failed to find needed subsystem <%s>",
				utils.FreeSWITCHAgent, utils.SessionS))
			return utils.ErrNotFound
		}
		sSIntConn := <-srvSessionS.GetIntenternalChan()
		srvSessionS.GetIntenternalChan() <- sSIntConn
		sS = utils.NewBiRPCInternalClient(sSIntConn.(*sessions.SessionS))
	} else {
		if sS, err = sp.NewConnection(utils.SessionS, sp.GetConfig().FsAgentCfg().SessionSConns); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.FreeSWITCHAgent, utils.SessionS, err.Error()))
			return
		}
	}
	if err = fS.Shutdown(); err != nil {
		return
	}
	fS.Lock()
	defer fS.Unlock()
	fS.fS.SetSessionSConnection(sS)
	fS.fS.Reload()
	if sSInternal { // bidirectional client backwards connection
		sS.(*utils.BiRPCInternalClient).SetClientConn(fS.fS)
		var rply string
		if err = sS.Call(utils.SessionSv1RegisterInternalBiJSONConn,
			utils.EmptyString, &rply); err != nil {
			utils.Logger.Crit(fmt.Sprintf("<%s> Could not connect to %s: %s",
				utils.FreeSWITCHAgent, utils.SessionS, err.Error()))
			return
		}
	}
	go func() {
		if err := fS.fS.Connect(); err != nil {
			utils.Logger.Err(fmt.Sprintf("<%s> error: %s!", utils.FreeSWITCHAgent, err))
			sp.GetExitChan() <- true // stop the engine here
		}
	}()
	return
}

// Shutdown stops the service
func (fS *FreeswitchAgent) Shutdown() (err error) {
	fS.Lock()
	defer fS.Unlock()
	if err = fS.fS.Shutdown(); err != nil {
		return
	}
	fS.fS = nil
	return
}

// GetRPCInterface returns the interface to register for server
func (fS *FreeswitchAgent) GetRPCInterface() interface{} {
	return fS.fS
}

// IsRunning returns if the service is running
func (fS *FreeswitchAgent) IsRunning() bool {
	fS.RLock()
	defer fS.RUnlock()
	return fS != nil && fS.fS != nil
}

// ServiceName returns the service name
func (fS *FreeswitchAgent) ServiceName() string {
	return utils.FreeSWITCHAgent
}