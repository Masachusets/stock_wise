package design

import (
	. "goa.design/goa/v3/dsl"
)

// ============================================================================
// API
// ============================================================================

var _ = API("stockwise", func() {
	Title("StockWise API")
	Description("Учёт материальных средств и оборудования")
	Version("1.0.0")

	Server("development", func() {
		Host("localhost", func() {
			URI("http://localhost:8080")
		})
	})
})

// ============================================================================
// Entity типы
// ============================================================================

var Nomenclature = Type("Nomenclature", func() {
	Description("Номенклатура")
	Attribute("code", String, "Код номенклатуры", func() {
		Example("08.01.02.00.00/0041")
	})
	Attribute("name", String, "Наименование", func() {
		Example("Ноутбук Clevo 15\"")
	})
	Required("code", "name")
})

var Card = Type("Card", func() {
	Description("Карточка сотрудника")
	Attribute("number", Int32, "Номер карточки (из КУ Ф-111)")
	Attribute("full_name", String, "ФИО (без звания)", func() {
		Example("Багликов А.С.")
	})
	Required("number", "full_name")
})

var Department = Type("Department", func() {
	Description("Подразделение")
	Attribute("code", Int32, "Код подразделения", func() {
		Example(100)
	})
	Attribute("type", String, "Тип подразделения", func() {
		Enum("warehouse", "upogg", "opk", "pogk", "pogz")
	})
	Attribute("name", String, "Наименование", func() {
		Example("СКЛАД")
	})
	Required("code", "type", "name")
})

var AssignmentInfo = Type("AssignmentInfo", func() {
	Description("Информация о закреплении оборудования")
	Attribute("target_type", String, "Тип закрепления", func() {
		Enum("employee", "department", "warehouse")
	})
	Attribute("card_number", Int32, "Номер карточки сотрудника")
	Attribute("full_name", String, "ФИО сотрудника", func() {
		Example("Багликов А.С.")
	})
	Attribute("dept_name", String, "Наименование подразделения")
	Attribute("operator_comment", String, "Комментарий оператора")
	Required("target_type")
})

var Equipment = Type("Equipment", func() {
	Description("Оборудование")
	Attribute("inventory_number", String, "Инвентарный номер", func() {
		Pattern("^ИТ\\d{5}$")
		Example("ИТ00205")
	})
	Attribute("serial_number", String, "Заводской номер")
	Attribute("nomenclature", Nomenclature, "Номенклатура")
	Attribute("model_name", String, "Наименование модели", func() {
		Example("Ноутбук Clevo 15\"")
	})
	Attribute("manufacture_date", String, "Дата изготовления", func() {
		Format(FormatDate)
	})
	Attribute("arrival_date", String, "Дата поступления", func() {
		Format(FormatDate)
	})
	Attribute("status", String, "Статус", func() {
		Enum("exp", "exp_int", "exp_sp", "broken", "written_off")
	})
	Attribute("form_number", String, "Номер формуляра")
	Attribute("location", String, "Место установки", func() {
		Example("219")
	})
	Attribute("notes", String, "Примечания")
	Attribute("assignment", AssignmentInfo, "Текущее закрепление")
	Required("inventory_number", "model_name", "status")
})

var WaybillsEquipment = Type("WaybillsEquipment", func() {
	Description("Позиция накладной (связь накладной и оборудования)")
	Attribute("waybill_id", Int32, "ID накладной")
	Attribute("equipment_id", Int32, "ID оборудования")
	Required("waybill_id", "equipment_id")
})

var Waybill = Type("Waybill", func() {
	Description("Накладная")
	Attribute("number", String, "Номер документа", func() {
		Example("Накладная №123")
	})
	Attribute("issue_date", String, "Дата документа", func() {
		Format(FormatDate)
	})
	Attribute("from_dept", Int32, "Код подразделения-отправителя")
	Attribute("to_dept", Int32, "Код подразделения-получателя")
	Attribute("status", String, "Статус", func() {
		Enum("draft", "signed", "archived")
	})
	Attribute("items", ArrayOf(WaybillsEquipment), "Позиции накладной")
	Required("number", "issue_date", "status")
})

var Assignment = Type("Assignment", func() {
	Description("Запись закрепления оборудования (история)")
	Attribute("target_type", String, "Тип закрепления", func() {
		Enum("employee", "department", "warehouse")
	})
	Attribute("card_number", Int32, "Номер карточки сотрудника")
	Attribute("full_name", String, "ФИО сотрудника")
	Attribute("dept_name", String, "Наименование подразделения")
	Attribute("assigned_at", String, "Дата закрепления", func() {
		Format(FormatDateTime)
	})
	Attribute("unassigned_at", String, "Дата снятия", func() {
		Format(FormatDateTime)
	})
	Attribute("operator_comment", String, "Комментарий оператора")
	Required("target_type", "assigned_at")
})

// ============================================================================
// Payload типы
// ============================================================================

var CreateCardPayload = Type("CreateCardPayload", func() {
	Description("Данные для создания карточки")
	Attribute("number", Int32, "Номер карточки")
	Attribute("full_name", String, "ФИО")
	Required("number", "full_name")
})

var UpdateCardPayload = Type("UpdateCardPayload", func() {
	Description("Данные для обновления карточки")
	Attribute("number", Int32, "Номер карточки")
	Attribute("full_name", String, "ФИО")
	Required("number")
})

var CreateEquipmentPayload = Type("CreateEquipmentPayload", func() {
	Description("Данные для создания оборудования")
	Attribute("inventory_number", String, "Инвентарный номер")
	Attribute("serial_number", String, "Заводской номер")
	Attribute("nomenclature_id", Int32, "ID номенклатуры")
	Attribute("model_name", String, "Наименование модели")
	Attribute("manufacture_date", String, "Дата изготовления", func() {
		Format(FormatDate)
	})
	Attribute("arrival_date", String, "Дата поступления", func() {
		Format(FormatDate)
	})
	Attribute("status", String, "Статус", func() {
		Enum("exp", "exp_int", "exp_sp", "broken", "written_off")
	})
	Attribute("form_number", String, "Номер формуляра")
	Attribute("location", String, "Место установки")
	Attribute("notes", String, "Примечания")
	Required("inventory_number", "model_name", "status")
})

var UpdateEquipmentPayload = Type("UpdateEquipmentPayload", func() {
	Description("Данные для обновления оборудования")
	Attribute("inventory_number", String, "Инвентарный номер")
	Attribute("serial_number", String, "Заводской номер")
	Attribute("nomenclature_id", Int32, "ID номенклатуры")
	Attribute("model_name", String, "Наименование модели")
	Attribute("manufacture_date", String, "Дата изготовления", func() {
		Format(FormatDate)
	})
	Attribute("arrival_date", String, "Дата поступления", func() {
		Format(FormatDate)
	})
	Attribute("status", String, "Статус", func() {
		Enum("exp", "exp_int", "exp_sp", "broken", "written_off")
	})
	Attribute("form_number", String, "Номер формуляра")
	Attribute("location", String, "Место установки")
	Attribute("notes", String, "Примечания")
	Required("inventory_number")
})

var CreateWaybillPayload = Type("CreateWaybillPayload", func() {
	Description("Данные для создания накладной")
	Attribute("number", String, "Номер документа")
	Attribute("issue_date", String, "Дата документа", func() {
		Format(FormatDate)
	})
	Attribute("from_dept", Int32, "Код подразделения-отправителя")
	Attribute("to_dept", Int32, "Код подразделения-получателя")
	Attribute("items", ArrayOf(WaybillsEquipment), "Позиции накладной")
	Required("number", "issue_date")
})

// ============================================================================
// List типы
// ============================================================================

var NomenclatureList = Type("NomenclatureList", func() {
	Attribute("nomenclatures", ArrayOf(Nomenclature))
	Required("nomenclatures")
})

var CardList = Type("CardList", func() {
	Attribute("cards", ArrayOf(Card))
	Required("cards")
})

var DepartmentList = Type("DepartmentList", func() {
	Attribute("departments", ArrayOf(Department))
	Required("departments")
})

var EquipmentList = Type("EquipmentList", func() {
	Attribute("equipments", ArrayOf(Equipment))
	Required("equipments")
})

var WaybillList = Type("WaybillList", func() {
	Attribute("waybills", ArrayOf(Waybill))
	Required("waybills")
})

var AssignmentList = Type("AssignmentList", func() {
	Attribute("assignments", ArrayOf(Assignment))
	Required("assignments")
})

// ============================================================================
// Сервисы
// ============================================================================

var _ = Service("nomenclatures", func() {
	Description("Справочник номенклатур (только чтение)")

	Error("not_found", String, "Номенклатура не найдена")

	Method("list", func() {
		Description("Список всех номенклатур")
		Result(NomenclatureList)
		HTTP(func() {
			GET("/dictionaries/nomenclatures")
			Response(StatusOK)
		})
	})

	Method("get", func() {
		Description("Получить номенклатуру по ID")
		Payload(func() {
			Attribute("id", Int32, "ID номенклатуры")
			Required("id")
		})
		Result(Nomenclature)
		Error("not_found")
		HTTP(func() {
			GET("/dictionaries/nomenclatures/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
})

var _ = Service("departments", func() {
	Description("Справочник подразделений (только чтение)")

	Error("not_found", String, "Подразделение не найдено")

	Method("list", func() {
		Description("Список всех подразделений")
		Result(DepartmentList)
		HTTP(func() {
			GET("/dictionaries/departments")
			Response(StatusOK)
		})
	})

	Method("get", func() {
		Description("Получить подразделение по коду")
		Payload(func() {
			Attribute("code", Int32, "Код подразделения")
			Required("code")
		})
		Result(Department)
		Error("not_found")
		HTTP(func() {
			GET("/dictionaries/departments/{code}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
})

var _ = Service("cards", func() {
	Description("Управление карточками сотрудников")

	Error("not_found", String, "Карточка не найдена")

	Method("list", func() {
		Description("Список всех карточек")
		Result(CardList)
		HTTP(func() {
			GET("/cards")
			Response(StatusOK)
		})
	})

	Method("get", func() {
		Description("Получить карточку по номеру")
		Payload(func() {
			Attribute("number", Int32, "Номер карточки")
			Required("number")
		})
		Result(Card)
		Error("not_found")
		HTTP(func() {
			GET("/cards/{number}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("create", func() {
		Description("Создать карточку")
		Payload(CreateCardPayload)
		Result(Card)
		HTTP(func() {
			POST("/cards")
			Response(StatusCreated)
		})
	})

	Method("update", func() {
		Description("Обновить карточку")
		Payload(UpdateCardPayload)
		Result(Card)
		Error("not_found")
		HTTP(func() {
			PUT("/cards/{number}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("delete", func() {
		Description("Удалить карточку")
		Payload(func() {
			Attribute("number", Int32, "Номер карточки")
			Required("number")
		})
		Error("not_found")
		HTTP(func() {
			DELETE("/cards/{number}")
			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
		})
	})
})

var _ = Service("equipments", func() {
	Description("Управление оборудованием")

	Error("not_found", String, "Оборудование не найдено")
	Error("conflict", String, "Инвентарный номер уже существует")

	Method("list", func() {
		Description("Список оборудования с фильтрацией")
		Payload(func() {
			Attribute("status", String, "Фильтр по статусу", func() {
				Enum("exp", "exp_int", "exp_sp", "broken", "written_off")
			})
			Attribute("nomenclature_id", Int32, "Фильтр по номенклатуре")
			Attribute("location", String, "Фильтр по местоположению")
			Attribute("search", String, "Поиск по наименованию/модели/серийнику")
		})
		Result(EquipmentList)
		HTTP(func() {
			GET("/equipments")
			Param("status")
			Param("nomenclature_id")
			Param("location")
			Param("search")
			Response(StatusOK)
		})
	})

	Method("get", func() {
		Description("Получить оборудование по инвентарному номеру (с закреплением)")
		Payload(func() {
			Attribute("inventory_number", String, "Инвентарный номер")
			Required("inventory_number")
		})
		Result(Equipment)
		Error("not_found")
		HTTP(func() {
			GET("/equipments/{inventory_number}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("create", func() {
		Description("Создать оборудование")
		Payload(CreateEquipmentPayload)
		Result(Equipment)
		Error("conflict")
		HTTP(func() {
			POST("/equipments")
			Response(StatusCreated)
			Response("conflict", StatusConflict)
		})
	})

	Method("update", func() {
		Description("Обновить оборудование по инвентарному номеру")
		Payload(UpdateEquipmentPayload)
		Result(Equipment)
		Error("not_found")
		Error("conflict")
		HTTP(func() {
			PUT("/equipments/{inventory_number}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
		})
	})

	Method("delete", func() {
		Description("Удалить оборудование (мягкое удаление)")
		Payload(func() {
			Attribute("inventory_number", String, "Инвентарный номер")
			Required("inventory_number")
		})
		Error("not_found")
		HTTP(func() {
			DELETE("/equipments/{inventory_number}")
			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
		})
	})
})

var _ = Service("waybills", func() {
	Description("Управление накладными")

	Error("not_found", String, "Накладная не найдена")
	Error("invalid_status", String, "Невозможно выполнить операцию для текущего статуса")

	Method("list", func() {
		Description("Список накладных")
		Result(WaybillList)
		HTTP(func() {
			GET("/waybills")
			Response(StatusOK)
		})
	})

	Method("get", func() {
		Description("Получить накладную по ID (с позициями)")
		Payload(func() {
			Attribute("id", Int32, "ID накладной")
			Required("id")
		})
		Result(Waybill)
		Error("not_found")
		HTTP(func() {
			GET("/waybills/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("create", func() {
		Description("Создать накладную (статус draft)")
		Payload(CreateWaybillPayload)
		Result(Waybill)
		HTTP(func() {
			POST("/waybills")
			Response(StatusCreated)
		})
	})

	Method("sign", func() {
		Description("Подписать накладную (draft → signed). Создаёт записи закреплений")
		Payload(func() {
			Attribute("id", Int32, "ID накладной")
			Required("id")
		})
		Result(Waybill)
		Error("not_found")
		Error("invalid_status")
		HTTP(func() {
			POST("/waybills/{id}/sign")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("invalid_status", StatusConflict)
		})
	})

	Method("archive", func() {
		Description("Архивировать накладную (signed → archived)")
		Payload(func() {
			Attribute("id", Int32, "ID накладной")
			Required("id")
		})
		Result(Waybill)
		Error("not_found")
		Error("invalid_status")
		HTTP(func() {
			POST("/waybills/{id}/archive")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("invalid_status", StatusConflict)
		})
	})

	Method("delete", func() {
		Description("Удалить накладную (только draft)")
		Payload(func() {
			Attribute("id", Int32, "ID накладной")
			Required("id")
		})
		Error("not_found")
		Error("invalid_status")
		HTTP(func() {
			DELETE("/waybills/{id}")
			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
			Response("invalid_status", StatusConflict)
		})
	})
})

var _ = Service("assignments", func() {
	Description("История закреплений оборудования")

	Error("not_found", String, "Закрепление не найдено")

	Method("list", func() {
		Description("Список закреплений с фильтрацией")
		Payload(func() {
			Attribute("equipment_id", Int32, "Фильтр по оборудованию")
			Attribute("is_active", Boolean, "Фильтр по активности")
		})
		Result(AssignmentList)
		HTTP(func() {
			GET("/assignments")
			Param("equipment_id")
			Param("is_active")
			Response(StatusOK)
		})
	})

	Method("get", func() {
		Description("Получить закрепление по ID")
		Payload(func() {
			Attribute("id", Int32, "ID закрепления")
			Required("id")
		})
		Result(Assignment)
		Error("not_found")
		HTTP(func() {
			GET("/assignments/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})
})
