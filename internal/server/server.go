package server

import (
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/shakil5281/peoplehub-api/docs"
	"github.com/shakil5281/peoplehub-api/internal/auth"
	"github.com/shakil5281/peoplehub-api/internal/config"
	"github.com/shakil5281/peoplehub-api/internal/database"
	"github.com/shakil5281/peoplehub-api/internal/handlers"
	"github.com/shakil5281/peoplehub-api/internal/middleware"
	"github.com/shakil5281/peoplehub-api/internal/repository"
	"github.com/shakil5281/peoplehub-api/internal/routes"
	"github.com/shakil5281/peoplehub-api/internal/service"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func New(cfg *config.Config) *gin.Engine {
	database.Connect(cfg)

	jwtCfg := auth.JWTConfig{
		Secret:          cfg.JWTSecret,
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "peoplehub-api",
	}

	userRepo := repository.NewUserRepository(database.DB)
	authRepo := repository.NewAuthRepository(database.DB)
	companyRepo := repository.NewCompanyRepository(database.DB)
	shiftRepo := repository.NewShiftRepository(database.DB)
	groupRepo := repository.NewGroupRepository(database.DB)
	floorRepo := repository.NewFloorRepository(database.DB)
	deptRepo := repository.NewDepartmentRepository(database.DB)
	sectionRepo := repository.NewSectionRepository(database.DB)
	desigRepo := repository.NewDesignationRepository(database.DB)
	lineRepo := repository.NewLineRepository(database.DB)
	divisionRepo := repository.NewDivisionRepository(database.DB)
	districtRepo := repository.NewDistrictRepository(database.DB)
	upazilaRepo := repository.NewUpazilaRepository(database.DB)
	unionRepo := repository.NewUnionRepository(database.DB)
	attendanceRepo := repository.NewAttendanceRepository(database.DB)
	dataLogRepo := repository.NewDataLogRepository(database.DB)
	employeeRepo := repository.NewEmployeeRepository(database.DB)
	requirementRepo := repository.NewRequirementRepository(database.DB)
	separationRepo := repository.NewSeparationRepository(database.DB)
	separationService := service.NewSeparationService(database.DB, separationRepo, employeeRepo, attendanceRepo)
	idCardRepo := repository.NewIdCardRepository(database.DB)

	authService := service.NewAuthService(userRepo, authRepo, jwtCfg)
	authHandler := handlers.NewAuthHandler(authService)
	employeeHandler := handlers.NewEmployeeHandler()
	companyHandler := handlers.NewCompanyHandler(companyRepo)
	shiftHandler := handlers.NewShiftHandler(shiftRepo)
	groupHandler := handlers.NewGroupHandler(groupRepo)
	floorHandler := handlers.NewFloorHandler(floorRepo)
	deptHandler := handlers.NewDepartmentHandler(deptRepo)
	sectionHandler := handlers.NewSectionHandler(sectionRepo)
	desigHandler := handlers.NewDesignationHandler(desigRepo)
	lineHandler := handlers.NewLineHandler(lineRepo)
	requirementHandler := handlers.NewRequirementHandler(requirementRepo)
	separationHandler := handlers.NewSeparationHandler(separationRepo, separationService)
	idCardHandler := handlers.NewIdCardHandler(idCardRepo)
	divisionHandler := handlers.NewDivisionHandler(divisionRepo)
	districtHandler := handlers.NewDistrictHandler(districtRepo)
	upazilaHandler := handlers.NewUpazilaHandler(upazilaRepo)
	unionHandler := handlers.NewUnionHandler(unionRepo)
	attendanceHandler := handlers.NewAttendanceHandler(attendanceRepo, employeeRepo, dataLogRepo, separationRepo)

	roleRepo := repository.NewRoleRepository(database.DB)
	roleHandler := handlers.NewRoleHandler(roleRepo)
	settingsRepo := repository.NewSettingsRepository(database.DB)
	settingsHandler := handlers.NewSettingsHandler(settingsRepo)
	userService := service.NewUserService(userRepo, authRepo, roleRepo)
	userHandler := handlers.NewUserHandler(userService)

	mdbReader := service.NewMDBReader()
	leaveRepo := repository.NewLeaveRepository(database.DB)
	tempShiftRepo := repository.NewTemporaryShiftRepository(database.DB)
	holidayRepo := repository.NewHolidayRepository(database.DB)
	dataLogService := service.NewDataLogService(dataLogRepo, mdbReader)
	attendanceProcessor := service.NewAttendanceProcessor(dataLogRepo, attendanceRepo, employeeRepo, shiftRepo, leaveRepo, tempShiftRepo, holidayRepo)
	dataLogHandler := handlers.NewDataLogHandler(dataLogRepo, dataLogService, attendanceProcessor)
	leaveHandler := handlers.NewLeaveHandler(leaveRepo, employeeRepo, attendanceRepo)
	salaryRepo := repository.NewSalaryRepository(database.DB)
	salaryService := service.NewSalaryService(employeeRepo, attendanceRepo, salaryRepo)
	salaryHandler := handlers.NewSalaryHandler(salaryService, salaryRepo)
	salaryIncrementRepo := repository.NewSalaryIncrementRepository(database.DB)
	salaryIncrementHandler := handlers.NewSalaryIncrementHandler(salaryIncrementRepo, employeeRepo)
	employeeImportHandler := handlers.NewEmployeeImportHandler(employeeRepo)
	orgImportHandler := handlers.NewOrganizationImportHandler()
	dashboardRepo := repository.NewDashboardRepository(database.DB)
	dashboardHandler := handlers.NewDashboardHandler(dashboardRepo)
	databaseHandler := handlers.NewDatabaseHandler(cfg)
	tempShiftHandler := handlers.NewTemporaryShiftHandler(tempShiftRepo, employeeRepo)
	punishmentRepo := repository.NewPunishmentRepository(database.DB)
	punishmentHandler := handlers.NewPunishmentHandler(punishmentRepo)
	dailyScheduleRepo := repository.NewDailyScheduleRepository(database.DB)
	dailyScheduleHandler := handlers.NewDailyScheduleHandler(dailyScheduleRepo)
	nightBillRepo := repository.NewNightBillRepository(database.DB)
	nightBillHandler := handlers.NewNightBillHandler(nightBillRepo, employeeRepo)
	tiffinBillRepo := repository.NewTiffinBillRepository(database.DB)
	tiffinBillHandler := handlers.NewTiffinBillHandler(tiffinBillRepo)
	holidayHandler := handlers.NewHolidayHandler(holidayRepo)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.Logger())
	r.Use(middleware.AuditMiddleware())

	// Serve uploaded files
	r.Static("/uploads", "./uploads")

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.Setup(r, authHandler, employeeHandler, companyHandler, shiftHandler, groupHandler, floorHandler, deptHandler, sectionHandler, desigHandler, lineHandler, orgImportHandler, dashboardHandler, databaseHandler, attendanceHandler, dataLogHandler, divisionHandler, districtHandler, upazilaHandler, unionHandler, requirementHandler, separationHandler, idCardHandler, leaveHandler, salaryHandler, salaryIncrementHandler, employeeImportHandler, tempShiftHandler, userHandler, roleHandler, settingsHandler, punishmentHandler, dailyScheduleHandler, nightBillHandler, tiffinBillHandler, holidayHandler, cfg.JWTSecret)

	return r
}
