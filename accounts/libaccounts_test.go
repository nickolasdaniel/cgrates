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

package accounts

import (
	"reflect"
	"testing"

	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
	"github.com/ericlagergren/decimal"
)

func TestCBDebitUnits(t *testing.T) {
	// with limit and unit factor
	cb := &concreteBalance{
		blnCfg: &utils.Balance{
			ID:   "TestCBDebitUnits",
			Type: utils.MetaConcrete,
			Opts: map[string]interface{}{
				utils.MetaBalanceLimit: decimal.New(-200, 0),
			},
			UnitFactors: []*utils.UnitFactor{
				{
					Factor: decimal.New(100, 0), // EuroCents
				},
			},
			Value: 500, // 500 EURcents
		},
		fltrS: new(engine.FilterS),
	}
	toDebit := decimal.New(6, 0)
	if dbted, uFctr, err := cb.debitUnits(toDebit, decimal.New(1, 0),
		&utils.CGREventWithOpts{CGREvent: &utils.CGREvent{Tenant: "cgrates.org"}}); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(cb.blnCfg.UnitFactors[0], uFctr) {
		t.Errorf("received unit factor: %+v", uFctr)
	} else if dbted.Cmp(toDebit) != 0 {
		t.Errorf("debited: %s", dbted)
	} else if cb.blnCfg.Value != -100.0 {
		t.Errorf("balance remaining: %f", cb.blnCfg.Value)
	}
	//with increment and not enough balance
	cb = &concreteBalance{
		blnCfg: &utils.Balance{
			ID:    "TestCBDebitUnits",
			Type:  utils.MetaConcrete,
			Value: 1.25,
		},
		fltrS: new(engine.FilterS),
	}
	if dbted, _, err := cb.debitUnits(
		new(decimal.Big).SetFloat64(1.5),
		new(decimal.Big).SetFloat64(0.1),
		&utils.CGREventWithOpts{CGREvent: &utils.CGREvent{Tenant: "cgrates.org"}}); err != nil {
		t.Error(err)
		// #FixMe: new(decimal.Big).SetFloat64(1.2) does not work :(
	} else if dbted.Cmp(decimal.New(12, 1)) != 0 { // only 1.2 is possible due to increment
		//} else if dbted.Cmp(new(decimal.Big).SetFloat64(1.2)) != 0 {
		t.Errorf("debited: %s, cmp: %v", dbted, dbted.Cmp(new(decimal.Big).SetFloat64(1.2)))
	} else if cb.blnCfg.Value != 0.05 {
		t.Errorf("balance remaining: %f", cb.blnCfg.Value)
	}
}