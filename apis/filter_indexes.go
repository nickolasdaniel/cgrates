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

package apis

import (
	"strings"
	"time"

	"github.com/cgrates/birpc/context"
	"github.com/cgrates/cgrates/engine"
	"github.com/cgrates/cgrates/utils"
)

type AttrGetFilterIndexes struct {
	Tenant      string
	Context     string
	ItemType    string
	FilterType  string
	FilterField string
	FilterValue string
	utils.Paginator
}

type AttrRemFilterIndexes struct {
	Tenant   string
	Context  string
	ItemType string
	APIOpts  map[string]interface{}
}

func (adms *AdminSv1) RemoveFilterIndexes(ctx *context.Context, arg *AttrRemFilterIndexes, reply *string) (err error) {
	if missing := utils.MissingStructFields(arg, []string{"ItemType"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	tntCtx := tnt
	switch arg.ItemType {
	case utils.MetaThresholds:
		arg.ItemType = utils.CacheThresholdFilterIndexes
	case utils.MetaRoutes:
		arg.ItemType = utils.CacheRouteFilterIndexes
	case utils.MetaStats:
		arg.ItemType = utils.CacheStatFilterIndexes
	case utils.MetaResources:
		arg.ItemType = utils.CacheResourceFilterIndexes
	case utils.MetaChargers:
		arg.ItemType = utils.CacheChargerFilterIndexes
	case utils.MetaAccounts:
		arg.ItemType = utils.CacheAccountsFilterIndexes
	case utils.MetaActions:
		arg.ItemType = utils.CacheActionProfilesFilterIndexes
	case utils.MetaRateProfiles:
		arg.ItemType = utils.CacheRateProfilesFilterIndexes
	case utils.MetaRateProfileRates:
		if missing := utils.MissingStructFields(arg, []string{"Context"}); len(missing) != 0 {
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.CacheRateFilterIndexes
		tntCtx = utils.ConcatenatedKey(tnt, arg.Context)
	case utils.MetaDispatchers:
		arg.ItemType = utils.CacheDispatcherFilterIndexes
	case utils.MetaAttributes:
		arg.ItemType = utils.CacheAttributeFilterIndexes
	}
	if err = adms.dm.RemoveIndexes(ctx, arg.ItemType, tntCtx, utils.EmptyString); err != nil {
		return
	}
	//generate a loadID for CacheFilterIndexes and store it in database
	if err := adms.dm.SetLoadIDs(ctx,
		map[string]int64{arg.ItemType: time.Now().UnixNano()}); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := adms.callCacheForRemoveIndexes(ctx, utils.IfaceAsString(arg.APIOpts[utils.CacheOpt]), arg.Tenant,
		arg.ItemType, []string{utils.MetaAny}, arg.APIOpts); err != nil {
		return utils.APIErrorHandler(err)
	}
	*reply = utils.OK
	return
}

func (adms *AdminSv1) GetFilterIndexes(ctx *context.Context, arg *AttrGetFilterIndexes, reply *[]string) (err error) {
	var indexes map[string]utils.StringSet
	var indexedSlice []string
	indexesFilter := make(map[string]utils.StringSet)
	if missing := utils.MissingStructFields(arg, []string{"ItemType"}); len(missing) != 0 { //Params missing
		return utils.NewErrMandatoryIeMissing(missing...)
	}
	tnt := arg.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	tntCtx := tnt
	switch arg.ItemType {
	case utils.MetaThresholds:
		arg.ItemType = utils.CacheThresholdFilterIndexes
	case utils.MetaRoutes:
		arg.ItemType = utils.CacheRouteFilterIndexes
	case utils.MetaStats:
		arg.ItemType = utils.CacheStatFilterIndexes
	case utils.MetaResources:
		arg.ItemType = utils.CacheResourceFilterIndexes
	case utils.MetaChargers:
		arg.ItemType = utils.CacheChargerFilterIndexes
	case utils.MetaAccounts:
		arg.ItemType = utils.CacheAccountsFilterIndexes
	case utils.MetaActions:
		arg.ItemType = utils.CacheActionProfilesFilterIndexes
	case utils.MetaRateProfiles:
		arg.ItemType = utils.CacheRateProfilesFilterIndexes
	case utils.MetaRateProfileRates:
		if missing := utils.MissingStructFields(arg, []string{"Context"}); len(missing) != 0 {
			return utils.NewErrMandatoryIeMissing(missing...)
		}
		arg.ItemType = utils.CacheRateFilterIndexes
		tntCtx = utils.ConcatenatedKey(tnt, arg.Context)
	case utils.MetaDispatchers:
		arg.ItemType = utils.CacheDispatcherFilterIndexes
	case utils.MetaAttributes:
		arg.ItemType = utils.CacheAttributeFilterIndexes
	}
	if indexes, err = adms.dm.GetIndexes(ctx,
		arg.ItemType, tntCtx, utils.EmptyString, true, true); err != nil {
		return
	}
	if arg.FilterType != utils.EmptyString {
		for val, strmap := range indexes {
			if strings.HasPrefix(val, arg.FilterType) {
				indexesFilter[val] = strmap
				for _, value := range strmap.AsSlice() {
					indexedSlice = append(indexedSlice, utils.ConcatenatedKey(val, value))
				}
			}
		}
		if len(indexedSlice) == 0 {
			return utils.ErrNotFound
		}
	}
	if arg.FilterField != utils.EmptyString {
		if len(indexedSlice) == 0 {
			indexesFilter = make(map[string]utils.StringSet)
			for val, strmap := range indexes {
				if strings.Index(val, arg.FilterField) != -1 {
					indexesFilter[val] = strmap
					for _, value := range strmap.AsSlice() {
						indexedSlice = append(indexedSlice, utils.ConcatenatedKey(val, value))
					}
				}
			}
			if len(indexedSlice) == 0 {
				return utils.ErrNotFound
			}
		} else {
			var cloneIndexSlice []string
			for val, strmap := range indexesFilter {
				if strings.Index(val, arg.FilterField) != -1 {
					for _, value := range strmap.AsSlice() {
						cloneIndexSlice = append(cloneIndexSlice, utils.ConcatenatedKey(val, value))
					}
				}
			}
			if len(cloneIndexSlice) == 0 {
				return utils.ErrNotFound
			}
			indexedSlice = cloneIndexSlice
		}
	}
	if arg.FilterValue != utils.EmptyString {
		if len(indexedSlice) == 0 {
			for val, strmap := range indexes {
				if strings.Index(val, arg.FilterValue) != -1 {
					for _, value := range strmap.AsSlice() {
						indexedSlice = append(indexedSlice, utils.ConcatenatedKey(val, value))
					}
				}
			}
			if len(indexedSlice) == 0 {
				return utils.ErrNotFound
			}
		} else {
			var cloneIndexSlice []string
			for val, strmap := range indexesFilter {
				if strings.Index(val, arg.FilterValue) != -1 {
					for _, value := range strmap.AsSlice() {
						cloneIndexSlice = append(cloneIndexSlice, utils.ConcatenatedKey(val, value))
					}
				}
			}
			if len(cloneIndexSlice) == 0 {
				return utils.ErrNotFound
			}
			indexedSlice = cloneIndexSlice
		}
	}
	if len(indexedSlice) == 0 {
		for val, strmap := range indexes {
			for _, value := range strmap.AsSlice() {
				indexedSlice = append(indexedSlice, utils.ConcatenatedKey(val, value))
			}
		}
	}
	if arg.Paginator.Limit != nil || arg.Paginator.Offset != nil {
		*reply = arg.Paginator.PaginateStringSlice(indexedSlice)
	} else {
		*reply = indexedSlice
	}
	return nil
}

// ComputeFilterIndexes selects which index filters to recompute
func (adms *AdminSv1) ComputeFilterIndexes(cntxt *context.Context, args *utils.ArgsComputeFilterIndexes, reply *string) (err error) {
	transactionID := utils.GenUUID()
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	cacheIDs := make(map[string][]string)

	var indexes utils.StringSet
	//ThresholdProfile Indexes
	if args.ThresholdS {
		cacheIDs[utils.ThresholdFilterIndexIDs] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheThresholdFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				th, e := adms.dm.GetThresholdProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(th.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.ThresholdS = indexes.Size() != 0
	}
	//StatQueueProfile Indexes
	if args.StatS {
		cacheIDs[utils.StatFilterIndexIDs] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheStatFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				sq, e := adms.dm.GetStatQueueProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(sq.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.StatS = indexes.Size() != 0
	}
	//ResourceProfile Indexes
	if args.ResourceS {
		cacheIDs[utils.ResourceFilterIndexIDs] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheResourceFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				rp, e := adms.dm.GetResourceProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(rp.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.ResourceS = indexes.Size() != 0
	}
	//RouteSProfile Indexes
	if args.RouteS {
		cacheIDs[utils.RouteFilterIndexIDs] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheRouteFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				rp, e := adms.dm.GetRouteProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(rp.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.RouteS = indexes.Size() != 0
	}
	//AttributeProfile Indexes
	if args.AttributeS {
		cacheIDs[utils.AttributeFilterIndexIDs] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheAttributeFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				ap, e := adms.dm.GetAttributeProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				fltrIDs := make([]string, len(ap.FilterIDs))
				for i, fltrID := range ap.FilterIDs {
					fltrIDs[i] = fltrID
				}

				return &fltrIDs, nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.AttributeS = indexes.Size() != 0
	}
	//ChargerProfile  Indexes
	if args.ChargerS {
		cacheIDs[utils.ChargerFilterIndexIDs] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheChargerFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				ap, e := adms.dm.GetChargerProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(ap.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.ChargerS = indexes.Size() != 0
	}
	//AccountFilter Indexes
	if args.AccountS {
		cacheIDs[utils.AccountsFilterIndexIDs] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheAccountsFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				acnts, e := adms.dm.GetAccount(cntxt, tnt, id)
				if e != nil {
					return nil, e
				}
				fltrIDs := make([]string, len(acnts.FilterIDs))
				for i, fltr := range acnts.FilterIDs {
					fltrIDs[i] = fltr
				}
				return &fltrIDs, nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.AccountS = indexes.Size() != 0
	}
	//ActionFilter Indexes
	if args.ActionS {
		cacheIDs[utils.ActionProfilesFilterIndexIDs] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheActionProfilesFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				act, e := adms.dm.GetActionProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				fltrIDs := make([]string, len(act.FilterIDs))
				for i, fltr := range act.FilterIDs {
					fltrIDs[i] = fltr
				}
				return &fltrIDs, nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.ActionS = indexes.Size() != 0
	}
	var ratePrf []string
	if args.RateS {
		cacheIDs[utils.RateProfilesFilterIndexIDs] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheRateProfilesFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				rtPrf, e := adms.dm.GetRateProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				ratePrf = append(ratePrf, utils.ConcatenatedKey(tnt, id))
				fltrIDs := make([]string, len(rtPrf.FilterIDs))
				for i, fltr := range rtPrf.FilterIDs {
					fltrIDs[i] = fltr
				}
				rtIds := make([]string, 0, len(rtPrf.Rates))

				for key := range rtPrf.Rates {
					rtIds = append(rtIds, key)
				}
				cacheIDs[utils.RateFilterIndexIDs] = rtIds
				_, e = engine.ComputeIndexes(cntxt, adms.dm, tnt, id, utils.CacheRateFilterIndexes,
					&rtIds, transactionID, func(_, id, _ string) (*[]string, error) {
						rateFilters := make([]string, len(rtPrf.Rates[id].FilterIDs))
						for i, fltrID := range rtPrf.Rates[id].FilterIDs {
							rateFilters[i] = fltrID
						}
						return &rateFilters, nil
					}, nil)
				if e != nil {
					return nil, e
				}
				return &fltrIDs, nil
			}, nil); err != nil {
			return utils.APIErrorHandler(err)
		}
		args.RateS = indexes.Size() != 0
	}
	//DispatcherProfile Indexes
	if args.DispatcherS {
		cacheIDs[utils.DispatcherFilterIndexIDs] = []string{utils.MetaAny}
		if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheDispatcherFilterIndexes,
			nil, transactionID, func(tnt, id, ctx string) (*[]string, error) {
				dsp, e := adms.dm.GetDispatcherProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
				if e != nil {
					return nil, e
				}
				return utils.SliceStringPointer(utils.CloneStringSlice(dsp.FilterIDs)), nil
			}, nil); err != nil && err != utils.ErrNotFound {
			return utils.APIErrorHandler(err)
		}
		args.DispatcherS = indexes.Size() != 0
	}

	//Now we move from tmpKey to the right key for each type
	//ThresholdProfile Indexes
	if args.ThresholdS {
		if err = adms.dm.SetIndexes(cntxt, utils.CacheThresholdFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//StatQueueProfile Indexes
	if args.StatS {
		if err = adms.dm.SetIndexes(cntxt, utils.CacheStatFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//ResourceProfile Indexes
	if args.ResourceS {
		if err = adms.dm.SetIndexes(cntxt, utils.CacheResourceFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//RouteProfile Indexes
	if args.RouteS {
		if err = adms.dm.SetIndexes(cntxt, utils.CacheRouteFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//AttributeProfile Indexes
	if args.AttributeS {
		if err = adms.dm.SetIndexes(cntxt, utils.CacheAttributeFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//ChargerProfile Indexes
	if args.ChargerS {
		if err = adms.dm.SetIndexes(cntxt, utils.CacheChargerFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//AccountProfile Indexes
	if args.AccountS {
		if err = adms.dm.SetIndexes(cntxt, utils.CacheAccountsFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return err
		}
	}
	//ActionProfile Indexes
	if args.ActionS {
		if err = adms.dm.SetIndexes(cntxt, utils.CacheActionProfilesFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return err
		}
	}
	//RateProfile Indexes
	if args.RateS {
		if err = adms.dm.SetIndexes(cntxt, utils.CacheRateProfilesFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return err
		}
		for _, tntId := range ratePrf {
			if err = adms.dm.SetIndexes(cntxt, utils.CacheRateFilterIndexes, tntId, nil, true, transactionID); err != nil {
				return err
			}

		}
	}
	//DispatcherProfile Indexes
	if args.DispatcherS {
		if err = adms.dm.SetIndexes(cntxt, utils.CacheDispatcherFilterIndexes, tnt, nil, true, transactionID); err != nil {
			return
		}
	}
	//generate a load
	//ID for CacheFilterIndexes and store it in database
	loadIDs := make(map[string]int64)
	timeNow := time.Now().UnixNano()
	for idx := range cacheIDs {
		loadIDs[utils.ArgCacheToInstance[idx]] = timeNow
	}
	if err := adms.dm.SetLoadIDs(cntxt, loadIDs); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := adms.callCacheForComputeIndexes(cntxt, utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, cacheIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}

// ComputeFilterIndexIDs computes specific filter indexes
func (adms *AdminSv1) ComputeFilterIndexIDs(cntxt *context.Context, args *utils.ArgsComputeFilterIndexIDs, reply *string) (err error) {
	transactionID := utils.NonTransactional
	tnt := args.Tenant
	if tnt == utils.EmptyString {
		tnt = adms.cfg.GeneralCfg().DefaultTenant
	}
	var indexes utils.StringSet
	cacheIDs := make(map[string][]string)
	//ThresholdProfile Indexes
	if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheThresholdFilterIndexes,
		&args.ThresholdIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			th, e := adms.dm.GetThresholdProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(th.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.ThresholdFilterIndexIDs] = indexes.AsSlice()
	}
	//StatQueueProfile Indexes
	if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheStatFilterIndexes,
		&args.StatIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			sq, e := adms.dm.GetStatQueueProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			cacheIDs[utils.StatFilterIndexIDs] = []string{sq.ID}
			return utils.SliceStringPointer(utils.CloneStringSlice(sq.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.StatFilterIndexIDs] = indexes.AsSlice()
	}
	//ResourceProfile Indexes
	if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheResourceFilterIndexes,
		&args.ResourceIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			rp, e := adms.dm.GetResourceProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			cacheIDs[utils.ResourceFilterIndexIDs] = []string{rp.ID}
			return utils.SliceStringPointer(utils.CloneStringSlice(rp.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.ResourceFilterIndexIDs] = indexes.AsSlice()
	}
	//RouteProfile Indexes
	if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheRouteFilterIndexes,
		&args.RouteIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			rp, e := adms.dm.GetRouteProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			cacheIDs[utils.RouteFilterIndexIDs] = []string{rp.ID}
			return utils.SliceStringPointer(utils.CloneStringSlice(rp.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.RouteFilterIndexIDs] = indexes.AsSlice()
	}
	//AttributeProfile Indexes
	if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheAttributeFilterIndexes,
		&args.AttributeIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			ap, e := adms.dm.GetAttributeProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(ap.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.AttributeFilterIndexIDs] = indexes.AsSlice()
	}
	//ChargerProfile  Indexes
	if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheChargerFilterIndexes,
		&args.ChargerIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			ap, e := adms.dm.GetChargerProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(ap.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.ChargerFilterIndexIDs] = indexes.AsSlice()
	}
	//AccountIndexes
	if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheAccountsFilterIndexes,
		&args.AccountIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			acc, e := adms.dm.GetAccount(cntxt, tnt, id)
			if e != nil {
				return nil, e
			}
			fltrIDs := make([]string, len(acc.FilterIDs))
			for i, fltr := range acc.FilterIDs {
				fltrIDs[i] = fltr
			}
			return &fltrIDs, nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.AccountsFilterIndexIDs] = indexes.AsSlice()
	}
	//ActionProfile Indexes
	if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheActionProfilesFilterIndexes,
		&args.ActionProfileIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			act, e := adms.dm.GetActionProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			fltrIDs := make([]string, len(act.FilterIDs))
			for i, fltr := range act.FilterIDs {
				fltrIDs[i] = fltr
			}
			return &fltrIDs, nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.ActionProfilesFilterIndexIDs] = indexes.AsSlice()
	}
	//RateProfile Indexes
	if _, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheRateProfilesFilterIndexes,
		&args.RateProfileIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			rpr, e := adms.dm.GetRateProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			fltrIDs := make([]string, len(rpr.FilterIDs))
			for i, fltrID := range rpr.FilterIDs {
				fltrIDs[i] = fltrID
			}

			rtIds := make([]string, 0, len(rpr.Rates))

			for key := range rpr.Rates {
				rtIds = append(rtIds, key)
			}
			indexesRate, e := engine.ComputeIndexes(cntxt, adms.dm, tnt, id, utils.CacheRateFilterIndexes,
				&rtIds, transactionID, func(_, id, _ string) (*[]string, error) {
					rateFilters := make([]string, len(rpr.Rates[id].FilterIDs))
					for i, fltrID := range rpr.Rates[id].FilterIDs {
						rateFilters[i] = fltrID
					}
					return &rateFilters, nil
				}, nil)
			if e != nil {
				return nil, e
			}
			if indexesRate.Size() != 0 {
				cacheIDs[utils.RateFilterIndexIDs] = indexesRate.AsSlice()
			}
			return &fltrIDs, nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.RateProfilesFilterIndexIDs] = indexes.AsSlice()
	}
	//DispatcherProfile Indexes
	if indexes, err = engine.ComputeIndexes(cntxt, adms.dm, tnt, utils.EmptyString, utils.CacheDispatcherFilterIndexes,
		&args.DispatcherIDs, transactionID, func(tnt, id, ctx string) (*[]string, error) {
			dsp, e := adms.dm.GetDispatcherProfile(cntxt, tnt, id, true, false, utils.NonTransactional)
			if e != nil {
				return nil, e
			}
			return utils.SliceStringPointer(utils.CloneStringSlice(dsp.FilterIDs)), nil
		}, nil); err != nil && err != utils.ErrNotFound {
		return utils.APIErrorHandler(err)
	}
	if indexes.Size() != 0 {
		cacheIDs[utils.DispatcherFilterIndexIDs] = indexes.AsSlice()
	}

	loadIDs := make(map[string]int64)
	timeNow := time.Now().UnixNano()
	for idx := range cacheIDs {
		loadIDs[utils.ArgCacheToInstance[idx]] = timeNow
	}
	if err := adms.dm.SetLoadIDs(cntxt, loadIDs); err != nil {
		return utils.APIErrorHandler(err)
	}
	if err := adms.callCacheForComputeIndexes(cntxt, utils.IfaceAsString(args.APIOpts[utils.CacheOpt]),
		args.Tenant, cacheIDs, args.APIOpts); err != nil {
		return err
	}
	*reply = utils.OK
	return nil
}