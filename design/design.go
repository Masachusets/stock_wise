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
	Attribute("id", Int32, "Идентификатор")
	Attribute("code", String, "Код номенклатуры", func() {
		Example("08.01.02.00.00/0041")
	})
	Attribute("name", String, "Наименование", func() {
		Example("Ноутбук Clevo 15\"")
	})
	Attribute("created_at", String, "Дата создания", func() {
		Format(FormatDate)
	})
	Required("id", "code", "name")
})

var Employee = Type("Employee", func() {
	Description("Сотрудник")
	Attribute("card_number", Int32, "Номер карточки (из КУ Ф-111)")
	Attribute("full_name", String, "ФИО (без звания)", func() {
		Example("Багликов А.С.")
	})
	Attribute("is_active", Boolean, "Активен", func() {
		Default(true)
	})
	Attribute("created_at", String, "Дата создания", func() {
		Format(FormatDate)
	})
	Required("card_number", "full_name")
})

var Department = Type("Department", func() {
	Description("Подразделение")
	Attribute("id", Int32, "Идентификатор")
	Attribute("code", String, "Код подразделения", func() {
		Example("02/1")
	})
	Attribute("name", String, "Наименование", func() {
		Example("Гомель ОПК")
	})
	Attribute("is_active", Boolean, "Активно", func() {
		Default(true)
	})
	Attribute("created_at", String, "Дата создания", func() {
		Format(FormatDate)
	})
	Required("id", "code", "name")
})

var Equipment = Type("Equipment", func() {
	Description("Оборудование")
	Attribute("id", Int32, "Идентификатор")
	Attribute("inventory_number", String, "Инвентарный номер", func() {
		Pattern("^ИТ\\d{5}$")
		Example("ИТ00205")
	})
	Attribute("serial_number", String, "Заводской номер")
	Attribute("mac_address", String, "MAC-адрес", func() {
		Format(FormatMAC)
	})
	Attribute("nomenclature_id", Int32, "ID номенклатуры")
	Attribute("model_name", String, "Наименование модели", func() {
		Example("Ноутбук Clevo 15\"")
	})
	Attribute("manufacture_year", Int32, "Год изготовления", func() {
		Minimum(1990)
		Maximum(2200)
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
	Attribute("nomenclature", Nomenclature, "Номенклатура (вложенная)")
	Attribute("created_at", String, "Дата создания", func() {
		Format(FormatDateTime)
	})
	Attribute("updated_at", String, "Дата обновления", func() {
		Format(FormatDateTime)
	})
	Required("id", "inventory_number", "model_name", "status")
})

var WaybillItem = Type("WaybillItem", func() {
	Description("Позиция накладной")
	Attribute("equipment_id", Int32, "ID оборудования")
	Attribute("equipment", Equipment, "Оборудование (вложенное)")
	Required("equipment_id")
})

var Waybill = Type("Waybill", func() {
	Description("Накладная")
	Attribute("id", Int32, "Идентификатор")
	Attribute("number", String, "Номер документа", func() {
		Example("Накладная №123")
	})
	Attribute("issue_date", String, "Дата документа", func() {
		Format(FormatDate)
	})
	Attribute("status", String, "Статус", func() {
		Enum("DRAFT", "SIGNED", "ARCHIVED")
	})
	Attribute("items", ArrayOf(WaybillItem), "Позиции накладной")
	Attribute("created_at", String, "Дата создания", func() {
		Format(FormatDateTime)
	})
	Required("id", "number", "issue_date", "status")
})

var Assignment = Type("Assignment", func() {
	Description("Закрепление оборудования")
	Attribute("id", Int32, "Идентификатор")
	Attribute("equipment_id", Int32, "ID оборудования")
	Attribute("target_type", String, "Тип закрепления", func() {
		Enum("employee", "department", "warehouse")
	})
	Attribute("employee_id", Int32, "Номер карточки сотрудника")
	Attribute("department_id", Int32, "ID подразделения")
	Attribute("waybill_id", Int32, "ID накладной")
	Attribute("assigned_at", String, "Дата закрепления", func() {
		Format(FormatDateTime)
	})
	Attribute("unassigned_at", String, "Дата снятия", func() {
		Format(FormatDateTime)
	})
	Attribute("is_active", Boolean, "Активно", func() {
		Default(true)
	})
	Attribute("operator_comment", String, "Комментарий оператора")
	Attribute("equipment", Equipment, "Оборудование (вложенное)")
	Attribute("employee", Employee, "Сотрудник (вложенный)")
	Attribute("department", Department, "Подразделение (вложенное)")
	Required("id", "equipment_id", "target_type", "is_active")
})

// ============================================================================
// Payload типы
// ============================================================================

var CreateNomenclaturePayload = Type("CreateNomenclaturePayload", func() {
	Description("Данные для создания номенклатуры")
	Attribute("code", String, "Код номенклатуры")
	Attribute("name", String, "Наименование")
	Required("code", "name")
})

var UpdateNomenclaturePayload = Type("UpdateNomenclaturePayload", func() {
	Description("Данные для обновления номенклатуры")
	Attribute("id", Int32, "Идентификатор")
	Attribute("code", String, "Код номенклатуры")
	Attribute("name", String, "Наименование")
	Required("id")
})

var CreateEmployeePayload = Type("CreateEmployeePayload", func() {
	Description("Данные для создания сотрудника")
	Attribute("card_number", Int32, "Номер карточки")
	Attribute("full_name", String, "ФИО")
	Required("card_number", "full_name")
})

var UpdateEmployeePayload = Type("UpdateEmployeePayload", func() {
	Description("Данные для обновления сотрудника")
	Attribute("card_number", Int32, "Номер карточки")
	Attribute("full_name", String, "ФИО")
	Attribute("is_active", Boolean, "Активен")
	Required("card_number")
})

var CreateDepartmentPayload = Type("CreateDepartmentPayload", func() {
	Description("Данные для создания подразделения")
	Attribute("code", String, "Код подразделения")
	Attribute("name", String, "Наименование")
	Required("code", "name")
})

var UpdateDepartmentPayload = Type("UpdateDepartmentPayload", func() {
	Description("Данные для обновления подразделения")
	Attribute("id", Int32, "Идентификатор")
	Attribute("code", String, "Код подразделения")
	Attribute("name", String, "Наименование")
	Attribute("is_active", Boolean, "Активно")
	Required("id")
})

var CreateEquipmentPayload = Type("CreateEquipmentPayload", func() {
	Description("Данные для создания оборудования")
	Attribute("inventory_number", String, "Инвентарный номер")
	Attribute("serial_number", String, "Заводской номер")
	Attribute("mac_address", String, "MAC-адрес")
	Attribute("nomenclature_id", Int32, "ID номенклатуры")
	Attribute("model_name", String, "Наименование модели")
	Attribute("manufacture_year", Int32, "Год изготовления")
	Attribute("arrival_date", String, "Дата поступления")
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
	Attribute("id", Int32, "Идентификатор")
	Attribute("inventory_number", String, "Инвентарный номер")
	Attribute("serial_number", String, "Заводской номер")
	Attribute("mac_address", String, "MAC-адрес")
	Attribute("nomenclature_id", Int32, "ID номенклатуры")
	Attribute("model_name", String, "Наименование модели")
	Attribute("manufacture_year", Int32, "Год изготовления")
	Attribute("arrival_date", String, "Дата поступления")
	Attribute("status", String, "Статус", func() {
		Enum("exp", "exp_int", "exp_sp", "broken", "written_off")
	})
	Attribute("form_number", String, "Номер формуляра")
	Attribute("location", String, "Место установки")
	Attribute("notes", String, "Примечания")
	Required("id")
})

var CreateWaybillPayload = Type("CreateWaybillPayload", func() {
	Description("Данные для создания накладной")
	Attribute("number", String, "Номер документа")
	Attribute("issue_date", String, "Дата документа")
	Attribute("items", ArrayOf(WaybillItem), "Позиции накладной")
	Required("number", "issue_date")
})

var NomenclatureList = Type("NomenclatureList", func() {
	Attribute("nomenclatures", ArrayOf(Nomenclature))
	Required("nomenclatures")
})

var EmployeeList = Type("EmployeeList", func() {
	Attribute("employees", ArrayOf(Employee))
	Required("employees")
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
// Services
// ============================================================================

var _ = Service("nomenclatures", func() {
	Description("Справочник номенклатур")

	Error("not_found", String, "Номенклатура не найдена")
	Error("conflict", String, "Код номенклатуры уже существует")

	Method("list", func() {
		Description("Список всех номенклатур")
		Result(NomenclatureList)
		HTTP(func() {
			GET("/nomenclatures")
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
			GET("/nomenclatures/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("create", func() {
		Description("Создать номенклатуру")
		Payload(CreateNomenclaturePayload)
		Result(Nomenclature)
		Error("conflict")
		HTTP(func() {
			POST("/nomenclatures")
			Response(StatusCreated)
			Response("conflict", StatusConflict)
		})
	})

	Method("update", func() {
		Description("Обновить номенклатуру")
		Payload(UpdateNomenclaturePayload)
		Result(Nomenclature)
		Error("not_found")
		Error("conflict")
		HTTP(func() {
			PUT("/nomenclatures/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
		})
	})

	Method("delete", func() {
		Description("Удалить номенклатуру")
		Payload(func() {
			Attribute("id", Int32, "ID номенклатуры")
			Required("id")
		})
		Error("not_found")
		HTTP(func() {
			DELETE("/nomenclatures/{id}")
			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
		})
	})
})

var _ = Service("employees", func() {
	Description("Справочник сотрудников")

	Error("not_found", String, "Сотрудник не найден")

	Method("list", func() {
		Description("Список всех сотрудников")
		Result(EmployeeList)
		HTTP(func() {
			GET("/employees")
			Response(StatusOK)
		})
	})

	Method("get", func() {
		Description("Получить сотрудника по номеру карточки")
		Payload(func() {
			Attribute("card_number", Int32, "Номер карточки")
			Required("card_number")
		})
		Result(Employee)
		Error("not_found")
		HTTP(func() {
			GET("/employees/{card_number}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("create", func() {
		Description("Создать сотрудника")
		Payload(CreateEmployeePayload)
		Result(Employee)
		HTTP(func() {
			POST("/employees")
			Response(StatusCreated)
		})
	})

	Method("update", func() {
		Description("Обновить сотрудника")
		Payload(UpdateEmployeePayload)
		Result(Employee)
		Error("not_found")
		HTTP(func() {
			PUT("/employees/{card_number}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("delete", func() {
		Description("Удалить сотрудника")
		Payload(func() {
			Attribute("card_number", Int32, "Номер карточки")
			Required("card_number")
		})
		Error("not_found")
		HTTP(func() {
			DELETE("/employees/{card_number}")
			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
		})
	})
})

var _ = Service("departments", func() {
	Description("Справочник подразделений")

	Error("not_found", String, "Подразделение не найдено")
	Error("conflict", String, "Код подразделения уже существует")

	Method("list", func() {
		Description("Список всех подразделений")
		Result(DepartmentList)
		HTTP(func() {
			GET("/departments")
			Response(StatusOK)
		})
	})

	Method("get", func() {
		Description("Получить подразделение по ID")
		Payload(func() {
			Attribute("id", Int32, "ID подразделения")
			Required("id")
		})
		Result(Department)
		Error("not_found")
		HTTP(func() {
			GET("/departments/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("create", func() {
		Description("Создать подразделение")
		Payload(CreateDepartmentPayload)
		Result(Department)
		Error("conflict")
		HTTP(func() {
			POST("/departments")
			Response(StatusCreated)
			Response("conflict", StatusConflict)
		})
	})

	Method("update", func() {
		Description("Обновить подразделение")
		Payload(UpdateDepartmentPayload)
		Result(Department)
		Error("not_found")
		Error("conflict")
		HTTP(func() {
			PUT("/departments/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
		})
	})

	Method("delete", func() {
		Description("Удалить подразделение")
		Payload(func() {
			Attribute("id", Int32, "ID подразделения")
			Required("id")
		})
		Error("not_found")
		HTTP(func() {
			DELETE("/departments/{id}")
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
		Description("Получить оборудование по ID")
		Payload(func() {
			Attribute("id", Int32, "ID оборудования")
			Required("id")
		})
		Result(Equipment)
		Error("not_found")
		HTTP(func() {
			GET("/equipments/{id}")
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
		Description("Обновить оборудование")
		Payload(UpdateEquipmentPayload)
		Result(Equipment)
		Error("not_found")
		Error("conflict")
		HTTP(func() {
			PUT("/equipments/{id}")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
			Response("conflict", StatusConflict)
		})
	})

	Method("delete", func() {
		Description("Удалить оборудование (soft delete)")
		Payload(func() {
			Attribute("id", Int32, "ID оборудования")
			Required("id")
		})
		Error("not_found")
		HTTP(func() {
			DELETE("/equipments/{id}")
			Response(StatusNoContent)
			Response("not_found", StatusNotFound)
		})
	})

	Method("getAssignment", func() {
		Description("Получить текущее закрепление оборудования")
		Payload(func() {
			Attribute("id", Int32, "ID оборудования")
			Required("id")
		})
		Result(Assignment)
		Error("not_found")
		HTTP(func() {
			GET("/equipments/{id}/assignment")
			Response(StatusOK)
			Response("not_found", StatusNotFound)
		})
	})

	Method("getAssignmentList", func() {
		Description("Получить историю закреплений оборудования")
		Payload(func() {
			Attribute("id", Int32, "ID оборудования")
			Required("id")
		})
		Result(AssignmentList)
		Error("not_found")
		HTTP(func() {
			GET("/equipments/{id}/assignments")
			Response(StatusOK)
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
		Description("Создать накладную (статус DRAFT)")
		Payload(CreateWaybillPayload)
		Result(Waybill)
		HTTP(func() {
			POST("/waybills")
			Response(StatusCreated)
		})
	})

	Method("sign", func() {
		Description("Подписать накладную (DRAFT → SIGNED). Создаёт записи в equipments_assignments")
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
		Description("Архивировать накладную (SIGNED → ARCHIVED)")
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
		Description("Удалить накладную (только DRAFT)")
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
