package main

import (
	assignments "github.com/Masachusets/stock_wise/gen/assignments"
	cards "github.com/Masachusets/stock_wise/gen/cards"
	departments "github.com/Masachusets/stock_wise/gen/departments"
	equipments "github.com/Masachusets/stock_wise/gen/equipments"
	nomenclatures "github.com/Masachusets/stock_wise/gen/nomenclatures"
	waybills "github.com/Masachusets/stock_wise/gen/waybills"
	svcassignments "github.com/Masachusets/stock_wise/internal/assignments"
	svccards "github.com/Masachusets/stock_wise/internal/cards"
	svcdepartments "github.com/Masachusets/stock_wise/internal/departments"
	svcequipments "github.com/Masachusets/stock_wise/internal/equipments"
	svcnomenclatures "github.com/Masachusets/stock_wise/internal/nomenclatures"
	svcwaybills "github.com/Masachusets/stock_wise/internal/waybills"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Services struct {
	Nomenclatures nomenclatures.Service
	Departments   departments.Service
	Cards         cards.Service
	Equipments    equipments.Service
	Waybills      waybills.Service
	WaybillsSvc   svcwaybills.WebService
	Assignments   assignments.Service
}

func NewServices(pool *pgxpool.Pool) *Services {
	wbRepo := svcwaybills.NewPostgresRepository(pool)
	wbSvc := svcwaybills.New(wbRepo)
	return &Services{
		Nomenclatures: svcnomenclatures.New(svcnomenclatures.NewPostgresRepository(pool)),
		Departments:   svcdepartments.New(svcdepartments.NewPostgresRepository(pool)),
		Cards:         svccards.New(svccards.NewPostgresRepository(pool)),
		Equipments:    svcequipments.New(svcequipments.NewPostgresRepository(pool)),
		Waybills:      wbSvc,
		WaybillsSvc:   wbSvc,
		Assignments:   svcassignments.New(svcassignments.NewPostgresRepository(pool)),
	}
}
