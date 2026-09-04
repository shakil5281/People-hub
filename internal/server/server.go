package server

import (
	"os"
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
	postOfficeRepo := repository.NewPostOfficeRepository(database.DB)
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
	postOfficeHandler := handlers.NewPostOfficeHandler(postOfficeRepo)
	missingAttRepo := repository.NewMissingAttendanceRepository(database.DB)
	attendanceHandler := handlers.NewAttendanceHandler(attendanceRepo, employeeRepo, dataLogRepo, separationRepo)
	missingAttendanceHandler := handlers.NewMissingAttendanceHandler(missingAttRepo, employeeRepo, attendanceRepo, companyRepo)

	roleRepo := repository.NewRoleRepository(database.DB)
	roleHandler := handlers.NewRoleHandler(roleRepo)
	settingsRepo := repository.NewSettingsRepository(database.DB)
	settingsHandler := handlers.NewSettingsHandler(settingsRepo)
	userService := service.NewUserService(userRepo, authRepo, roleRepo)
	userHandler := handlers.NewUserHandler(userService)

	mdbReader := service.NewMDBReader()
	leaveRepo := repository.NewLeaveRepository(database.DB)
	tempShiftRepo := repository.NewTemporaryShiftRepository(database.DB)
	rosterRepo := repository.NewRosterRepository(database.DB)
	holidayRepo := repository.NewHolidayRepository(database.DB)
	dataLogService := service.NewDataLogService(dataLogRepo, mdbReader)
	attendanceProcessor := service.NewAttendanceProcessor(dataLogRepo, attendanceRepo, employeeRepo, shiftRepo, leaveRepo, tempShiftRepo, rosterRepo, holidayRepo, missingAttRepo)
	attendanceProcessor.SetSeparationRepo(separationRepo)
	dataLogHandler := handlers.NewDataLogHandler(dataLogRepo, dataLogService, attendanceProcessor)
	leaveHandler := handlers.NewLeaveHandler(leaveRepo, employeeRepo, attendanceRepo)
	salaryRepo := repository.NewSalaryRepository(database.DB)
	salaryIncrementRepo := repository.NewSalaryIncrementRepository(database.DB)
	advanceSalaryRepo := repository.NewAdvanceSalaryRepository(database.DB)
	
	otEarlyExitRepo := repository.NewOtEarlyExitRepository(database.DB)
	otEarlyExitService := service.NewOtEarlyExitService(otEarlyExitRepo, holidayRepo)
	salaryService := service.NewSalaryService(employeeRepo, attendanceRepo, salaryRepo, groupRepo, otEarlyExitRepo, otEarlyExitService, advanceSalaryRepo, separationRepo)
	salaryService.SetIncrementRepo(salaryIncrementRepo)
	salaryHandler := handlers.NewSalaryHandler(salaryService, salaryRepo)

	// Earned Leave module
	earnedLeavePolicyRepo := repository.NewEarnedLeavePolicyRepository(database.DB)
	earnedLeaveLedgerRepo := repository.NewEarnedLeaveLedgerRepository(database.DB)
	earnedLeaveBalanceRepo := repository.NewEarnedLeaveBalanceRepository(database.DB)
	earnedLeaveSalaryRepo := repository.NewEarnedLeaveSalaryRepository(database.DB)
	earnedLeaveService := service.NewEarnedLeaveService(earnedLeavePolicyRepo, earnedLeaveLedgerRepo, earnedLeaveBalanceRepo, earnedLeaveSalaryRepo, employeeRepo, separationRepo, leaveRepo, database.DB)
	earnedLeaveHandler := handlers.NewEarnedLeaveHandler(earnedLeaveService, earnedLeavePolicyRepo, earnedLeaveLedgerRepo, earnedLeaveSalaryRepo)
	otEarlyExitHandler := handlers.NewOtEarlyExitHandler(otEarlyExitRepo, otEarlyExitService)
	
	salaryIncrementHandler := handlers.NewSalaryIncrementHandler(salaryIncrementRepo, employeeRepo)
	advanceSalaryHandler := handlers.NewAdvanceSalaryHandler(advanceSalaryRepo, employeeRepo)
	employeeImportHandler := handlers.NewEmployeeImportHandler(employeeRepo)
	orgImportHandler := handlers.NewOrganizationImportHandler()
	dashboardRepo := repository.NewDashboardRepository(database.DB)
	dashboardHandler := handlers.NewDashboardHandler(dashboardRepo)
	databaseHandler := handlers.NewDatabaseHandler(cfg)
	tempShiftHandler := handlers.NewTemporaryShiftHandler(tempShiftRepo, employeeRepo)
	rosterHandler := handlers.NewRosterHandler(rosterRepo, employeeRepo)
	punishmentRepo := repository.NewPunishmentRepository(database.DB)
	punishmentHandler := handlers.NewPunishmentHandler(punishmentRepo)
	dailyScheduleRepo := repository.NewDailyScheduleRepository(database.DB)
	dailyScheduleHandler := handlers.NewDailyScheduleHandler(dailyScheduleRepo)
	tiffinBillRepo := repository.NewTiffinBillRepository(database.DB)
	tiffinBillHandler := handlers.NewTiffinBillHandler(tiffinBillRepo)
	holidayHandler := handlers.NewHolidayHandler(holidayRepo, shiftRepo, attendanceProcessor)
	systemLogRepo := repository.NewSystemLogRepository(database.DB)
	systemLogHandler := handlers.NewSystemLogHandler(systemLogRepo)

	notificationRepo := repository.NewNotificationRepository(database.DB)
	notificationHandler := handlers.NewNotificationHandler(notificationRepo)

	notificationChecker := service.NewNotificationChecker(database.DB, notificationRepo, employeeRepo)
	notificationChecker.Start(2 * time.Hour)

	eidBonusRepo := repository.NewEidBonusRepository(database.DB)
	eidBonusService := service.NewEidBonusService(employeeRepo, eidBonusRepo)
	eidBonusHandler := handlers.NewEidBonusHandler(eidBonusService, eidBonusRepo)

	nightBillRepo := repository.NewNightBillRepository(database.DB)
	nightBillEmployeeListRepo := repository.NewNightBillEmployeeListRepository(database.DB)
	nightBillService := service.NewNightBillService(database.DB, nightBillRepo, nightBillEmployeeListRepo)
	nightBillHandler := handlers.NewNightBillHandler(nightBillService, nightBillRepo, companyRepo)

	nightBillEmployeeListHandler := handlers.NewNightBillEmployeeListHandler(nightBillEmployeeListRepo)

	migrationRepo := repository.NewMigrationRepository(database.DB)
	migrationService := service.NewMigrationService(database.DB, migrationRepo, employeeRepo)
	migrationHandler := handlers.NewMigrationHandler(migrationRepo, migrationService)
	zktecoSyncHandler := handlers.NewZKTecoSyncHandler(employeeRepo)
	salaryAccountImportHandler := handlers.NewSalaryAccountImportHandler()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.Logger())
	r.Use(middleware.AuditMiddleware())
	// Global rate limit: 300 req/min per IP — 100 was too low for dashboard + header concurrent fetches (4 parallel = 960/min flood)
	r.Use(middleware.RateLimit(300, time.Minute))

	// Serve uploaded files - PROTECTED: require auth, no anonymous enumeration
	// Old: r.Static("/uploads", "./uploads") - REMOVED for security
	uploads := r.Group("/uploads")
	uploads.Use(middleware.AuthMiddleware(jwtCfg.Secret))
	uploads.GET("/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		// Prevent path traversal - gin already cleans, but double check
		if filepath == "" {
			c.JSON(404, gin.H{"error": "file not found"})
			return
		}
		c.File("./uploads" + filepath)
	})

	// Swagger UI - protected in production, open in development
	if os.Getenv("GO_ENV") == "production" {
		swagger := r.Group("/swagger")
		swagger.Use(middleware.AuthMiddleware(jwtCfg.Secret), middleware.RequireRole("super_admin"))
		swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	} else {
		r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	routes.Setup(r, authHandler, employeeHandler, companyHandler, shiftHandler, groupHandler, floorHandler, deptHandler, sectionHandler, desigHandler, lineHandler, orgImportHandler, dashboardHandler, databaseHandler, attendanceHandler, dataLogHandler, divisionHandler, districtHandler, upazilaHandler, unionHandler, postOfficeHandler, requirementHandler, separationHandler, idCardHandler, leaveHandler, salaryHandler, salaryIncrementHandler, advanceSalaryHandler, eidBonusHandler, employeeImportHandler, tempShiftHandler, rosterHandler, userHandler, roleHandler, settingsHandler, punishmentHandler, dailyScheduleHandler, tiffinBillHandler, holidayHandler, systemLogHandler, notificationHandler, missingAttendanceHandler, otEarlyExitHandler, nightBillHandler, nightBillEmployeeListHandler, migrationHandler, zktecoSyncHandler, salaryAccountImportHandler, earnedLeaveHandler, cfg.JWTSecret)

	return r
}
