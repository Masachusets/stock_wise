package api

import (
	"context"
	"net/http"

	assignments "github.com/Masachusets/stock_wise/gen/assignments"
	cards "github.com/Masachusets/stock_wise/gen/cards"
	departments "github.com/Masachusets/stock_wise/gen/departments"
	equipments "github.com/Masachusets/stock_wise/gen/equipments"
	assignmentssvr "github.com/Masachusets/stock_wise/gen/http/assignments/server"
	cardssvr "github.com/Masachusets/stock_wise/gen/http/cards/server"
	departmentssvr "github.com/Masachusets/stock_wise/gen/http/departments/server"
	equipmentssvr "github.com/Masachusets/stock_wise/gen/http/equipments/server"
	nomenclaturessvr "github.com/Masachusets/stock_wise/gen/http/nomenclatures/server"
	waybillssvr "github.com/Masachusets/stock_wise/gen/http/waybills/server"
	nomenclatures "github.com/Masachusets/stock_wise/gen/nomenclatures"
	waybills "github.com/Masachusets/stock_wise/gen/waybills"
	"goa.design/clue/log"
	goahttp "goa.design/goa/v3/http"
)

func RegisterRoutes(mux goahttp.Muxer, nomenclaturesSvc nomenclatures.Service, departmentsSvc departments.Service, cardsSvc cards.Service, equipmentsSvc equipments.Service, waybillsSvc waybills.Service, assignmentsSvc assignments.Service) {
	dec := goahttp.RequestDecoder
	enc := goahttp.ResponseEncoder
	eh := errorHandler()

	nomEndpoints := nomenclatures.NewEndpoints(nomenclaturesSvc)
	deptEndpoints := departments.NewEndpoints(departmentsSvc)
	cardsEndpoints := cards.NewEndpoints(cardsSvc)
	eqEndpoints := equipments.NewEndpoints(equipmentsSvc)
	wbEndpoints := waybills.NewEndpoints(waybillsSvc)
	asnEndpoints := assignments.NewEndpoints(assignmentsSvc)

	nomSvr := nomenclaturessvr.New(nomEndpoints, mux, dec, enc, eh, nil)
	deptSvr := departmentssvr.New(deptEndpoints, mux, dec, enc, eh, nil)
	cardsSvr := cardssvr.New(cardsEndpoints, mux, dec, enc, eh, nil)
	eqSvr := equipmentssvr.New(eqEndpoints, mux, dec, enc, eh, nil)
	wbSvr := waybillssvr.New(wbEndpoints, mux, dec, enc, eh, nil)
	asnSvr := assignmentssvr.New(asnEndpoints, mux, dec, enc, eh, nil)

	nomenclaturessvr.Mount(mux, nomSvr)
	departmentssvr.Mount(mux, deptSvr)
	cardssvr.Mount(mux, cardsSvr)
	equipmentssvr.Mount(mux, eqSvr)
	waybillssvr.Mount(mux, wbSvr)
	assignmentssvr.Mount(mux, asnSvr)

	log.Printf(context.Background(), "api routes registered")
}

func errorHandler() func(context.Context, http.ResponseWriter, error) {
	return func(ctx context.Context, w http.ResponseWriter, err error) {
		log.Errorf(ctx, err, "api error: %v", err)
	}
}
