package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type StringArray []string

func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = []string{}
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan StringArray: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, sa)
}

func (sa StringArray) Value() (driver.Value, error) {
	if len(sa) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal(sa)
}

var db *gorm.DB
var clients = make(map[int]*websocket.Conn)
var r *gin.Engine

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type User struct {
	ID            int             `json:"id" gorm:"primaryKey"`
	Username      string          `json:"username" gorm:"unique;not null"`
	Email         string          `json:"email" gorm:"unique;not null"`
	Password      string          `json:"password" gorm:"not null"`
	Role          string          `json:"role" gorm:"not null"`
	FullName      string          `json:"full_name"`
	Description   string          `json:"description"`
	AvatarURL     string          `json:"avatar_url"`
	Services      StringArray     `json:"services" gorm:"type:jsonb"`
	Balance       decimal.Decimal `json:"balance" gorm:"type:decimal(10,2);default:0"`
	EncryptedCard string          `json:"encrypted_card"`
	PayoutAmount  decimal.Decimal `json:"payout_amount" gorm:"type:decimal(10,2);default:0"`
	CreatedAt     time.Time       `json:"created_at" gorm:"autoCreateTime"`
}

type Course struct {
	ID          int             `json:"id" gorm:"primaryKey"`
	TeacherID   int             `json:"teacher_id"`
	Title       string          `json:"title" gorm:"not null"`
	Services    StringArray     `json:"services" gorm:"type:jsonb"`
	Description string          `json:"description"`
	NetPrice    decimal.Decimal `json:"net_price" gorm:"type:decimal(10,2)"`
	GrossPrice  decimal.Decimal `json:"gross_price" gorm:"type:decimal(10,2)"`
	VideoURL    string          `json:"video_url"`
	CreatedAt   time.Time       `json:"created_at" gorm:"autoCreateTime"`
	Teacher     User            `json:"teacher" gorm:"foreignKey:TeacherID"`
}

type Enrollment struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	CourseID  int       `json:"course_id"`
	UserID    int       `json:"user_id"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	Course    Course    `json:"course" gorm:"foreignKey:CourseID"`
}

type Review struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	AuthorID  int       `json:"author_id"`
	CourseID  int       `json:"course_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	Author    User      `json:"author" gorm:"foreignKey:AuthorID"`
}

type Payment struct {
	ID            int             `json:"id" gorm:"primaryKey"`
	UserID        int             `json:"user_id"`
	CourseID      int             `json:"course_id"`
	GrossAmount   decimal.Decimal `json:"gross_amount" gorm:"type:decimal(10,2)"`
	Commission    decimal.Decimal `json:"commission" gorm:"type:decimal(10,2)"`
	NetAmount     decimal.Decimal `json:"net_amount" gorm:"type:decimal(10,2)"`
	Status        string          `json:"status" gorm:"default:'pending'"`
	YookassaID    string          `json:"yookassa_id"`
	TransactionID string          `json:"transaction_id"`
	CreatedAt     time.Time       `json:"created_at" gorm:"autoCreateTime"`
}

type Message struct {
	ID         int        `json:"id" gorm:"primaryKey"`
	SenderID   int        `json:"sender_id"`
	ReceiverID int        `json:"receiver_id"`
	Content    string     `json:"content"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	ReadAt     *time.Time `json:"read_at"`
	Sender     User       `json:"sender" gorm:"foreignKey:SenderID"`
	Receiver   User       `json:"receiver" gorm:"foreignKey:ReceiverID"`
}

type Notification struct {
	ID        int        `json:"id" gorm:"primaryKey"`
	UserID    int        `json:"user_id"`
	Type      string     `json:"type"`
	Content   string     `json:"content"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	ReadAt    *time.Time `json:"read_at"`
}

type Dialog struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	Sender     User      `json:"sender" gorm:"foreignKey:SenderID"`
	Receiver   User      `json:"receiver" gorm:"foreignKey:ReceiverID"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки .env файла: %v", err)
	}
	dsn := os.Getenv("DSN")
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	log.Println("Database connected successfully")
	if err := db.AutoMigrate(&User{}, &Course{}, &Enrollment{}, &Review{}, &Payment{}, &Message{}, &Notification{}, &Dialog{}); err != nil {
		log.Fatalf("Ошибка миграции БД: %v", err)
	}
	log.Println("Database migration completed")

	// Проверка и создание тестового нутрициолога
	var count int64
	if err := db.Model(&User{}).Where("role = ?", "nutri").Count(&count).Error; err != nil {
		log.Printf("Ошибка проверки тестового нутрициолога: %v", err)
	}
	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("test123456"), bcrypt.DefaultCost)
		testNutri := User{
			Username:    "testnutri",
			Email:       "testnutri@example.com",
			Password:    string(hashedPassword),
			Role:        "nutri",
			FullName:    "Тестовый Нутрициолог",
			Description: "Тестовое описание услуг",
			Services:    StringArray{"Диета", "Консультации"},
		}
		if err := db.Create(&testNutri).Error; err != nil {
			log.Printf("Ошибка создания тестового нутрициолога: %v", err)
		} else {
			log.Println("Создан тестовый нутрициолог")
		}
	}

	// Проверка и создание учетной записи администратора
	var adminCount int64
	if err := db.Model(&User{}).Where("username = ?", "adminis").Count(&adminCount).Error; err != nil {
		log.Printf("Ошибка проверки администратора: %v", err)
	}
	if adminCount == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Cl33l2l4jswi98"), bcrypt.DefaultCost)
		admin := User{
			Username: "adminis",
			Email:    "admin@example.com",
			Password: string(hashedPassword),
			Role:     "admin",
			FullName: "Администратор",
		}
		if err := db.Create(&admin).Error; err != nil {
			log.Printf("Ошибка создания администратора: %v", err)
		} else {
			log.Println("Создан администратор")
		}
	}

	r = gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
	r.Static("/avatars", "./Uploads/avatars")
	api := r.Group("/api")
	api.POST("/register", register)
	api.POST("/login", login)
	api.GET("/profile", authMiddleware, getProfile)
	api.GET("/profile/:id", getProfile)
	api.PUT("/profile", authMiddleware, updateProfile)
	api.POST("/profile/update-card", authMiddleware, updateCard)
	api.POST("/profile/upload-avatar", authMiddleware, uploadAvatarHandler)
	api.GET("/search", searchCourses)
	api.GET("/courses", authMiddleware, getCourses)
	api.POST("/courses", authMiddleware, createCourse)
	api.GET("/courses/:id", getCourse)
	api.POST("/payments/create", authMiddleware, createPayment)
	api.GET("/payments/return", authMiddleware, returnPayment)
	api.POST("/webhook/yookassa", webhookYookassa)
	api.GET("/reviews/user/:id", getUserReviews)
	api.GET("/reviews/course/:id", getCourseReviews)
	api.GET("/reviews/random", getRandomReviews)
	api.POST("/reviews", authMiddleware, createReview)
	api.GET("/enrolled", authMiddleware, getEnrolled)
	api.GET("/nutris", getNutris)
	api.POST("/start-chat", authMiddleware, startChat)
	api.GET("/chats", authMiddleware, getChats)
	api.GET("/messages", authMiddleware, getMessages)
	api.POST("/messages", authMiddleware, sendMessage)
	api.PUT("/messages/read", authMiddleware, markRead)
	api.GET("/admin/nutris", authMiddleware, getAdminNutris)
	api.POST("/admin/decrypt-card", authMiddleware, decryptCard)
	api.POST("/admin/payout", authMiddleware, processPayout)
	api.POST("/admin/update-payout-amount", authMiddleware, updatePayoutAmount)
	r.GET("/ws", handleWebSocket)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}

func authMiddleware(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		log.Println("authMiddleware: Токен не предоставлен")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Токен не предоставлен"})
		c.Abort()
		return
	}
	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		log.Printf("authMiddleware: Неверный токен: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
		c.Abort()
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("authMiddleware: Неверные данные токена")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные данные токена"})
		c.Abort()
		return
	}
	userID := int(claims["id"].(float64))
	role := claims["role"].(string)
	c.Set("userID", userID)
	c.Set("role", role)
	c.Next()
}

func register(c *gin.Context) {
	var input struct {
		Username    string `json:"username"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		Role        string `json:"role"`
		FullName    string `json:"full_name"`
		Description string `json:"description"`
	}
	if err := c.BindJSON(&input); err != nil {
		log.Printf("register: Неверные данные: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if input.Role != "client" && input.Role != "nutri" {
		log.Println("register: Неверная роль")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверная роль"})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("register: Ошибка хеширования пароля: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка хеширования пароля"})
		return
	}
	user := User{
		Username:    input.Username,
		Email:       input.Email,
		Password:    string(hashedPassword),
		Role:        input.Role,
		FullName:    input.FullName,
		Description: input.Description,
	}
	if err := db.Create(&user).Error; err != nil {
		log.Printf("register: Ошибка создания пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания пользователя"})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		log.Printf("register: Ошибка генерации токена: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString, "role": user.Role, "id": user.ID})
}

func login(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&input); err != nil {
		log.Printf("login: Неверные данные: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	var user User
	if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		log.Printf("login: Пользователь не найден: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные данные"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		log.Println("login: Неверный пароль")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные данные"})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		log.Printf("login: Ошибка генерации токена: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString, "role": user.Role, "id": user.ID})
}

func getProfile(c *gin.Context) {
	idStr := c.Param("id")
	var userID int
	var err error
	if idStr != "" {
		userID, err = strconv.Atoi(idStr)
		if err != nil {
			log.Printf("getProfile: Неверный ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
			return
		}
	} else {
		userID = c.GetInt("userID")
	}
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		log.Printf("getProfile: Пользователь не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}
	var courses []Course
	if err := db.Where("teacher_id = ?", userID).Find(&courses).Error; err != nil {
		log.Printf("getProfile: Ошибка получения курсов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения курсов"})
		return
	}
	var reviews []Review
	if err := db.Preload("Author").Where("course_id IN (SELECT id FROM courses WHERE teacher_id = ?)", userID).Find(&reviews).Error; err != nil {
		log.Printf("getProfile: Ошибка получения отзывов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения отзывов"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"profile": user, "courses": courses, "reviews": reviews})
}

func updateProfile(c *gin.Context) {
	userID := c.GetInt("userID")
	var input struct {
		FullName    string `json:"full_name"`
		Description string `json:"description"`
	}
	if err := c.BindJSON(&input); err != nil {
		log.Printf("updateProfile: Неверные данные: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if err := db.Model(&User{}).Where("id = ?", userID).Updates(User{FullName: input.FullName, Description: input.Description}).Error; err != nil {
		log.Printf("updateProfile: Ошибка обновления профиля: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления профиля"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Профиль обновлен"})
}

func updateCard(c *gin.Context) {
	userID := c.GetInt("userID")
	var input struct {
		CardNumber string `json:"card_number"`
	}
	if err := c.BindJSON(&input); err != nil {
		log.Printf("updateCard: Неверные данные: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if matched, _ := regexp.MatchString(`^\d{16}$`, input.CardNumber); !matched {
		log.Println("updateCard: Неверный номер карты")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный номер карты"})
		return
	}
	key, err := hex.DecodeString(os.Getenv("AES_KEY"))
	if err != nil {
		log.Printf("updateCard: Ошибка декодирования ключа AES: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка шифрования"})
		return
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("updateCard: Ошибка создания шифра: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка шифрования"})
		return
	}
	plaintext := []byte(input.CardNumber)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Printf("updateCard: Ошибка генерации IV: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка шифрования"})
		return
	}
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)
	encrypted := hex.EncodeToString(ciphertext)
	if err := db.Model(&User{}).Where("id = ?", userID).Update("encrypted_card", encrypted).Error; err != nil {
		log.Printf("updateCard: Ошибка обновления карты: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления карты"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Карта обновлена"})
}

func uploadAvatarHandler(c *gin.Context) {
	userID := c.GetInt("userID")
	file, err := c.FormFile("avatar")
	if err != nil {
		log.Printf("uploadAvatar: Ошибка получения файла: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка загрузки файла"})
		return
	}
	if file.Size > 5*1024*1024 {
		log.Println("uploadAvatar: Файл слишком большой")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл превышает 5MB"})
		return
	}
	filename := uuid.New().String() + ".jpg"
	path := "./Uploads/avatars/" + filename
	if err := os.MkdirAll("./Uploads/avatars", os.ModePerm); err != nil {
		log.Printf("uploadAvatar: Ошибка создания директории: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}
	if err := c.SaveUploadedFile(file, path); err != nil {
		log.Printf("uploadAvatar: Ошибка сохранения файла: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения файла"})
		return
	}
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		log.Printf("uploadAvatar: Пользователь не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}
	user.AvatarURL = "/avatars/" + filename
	if err := db.Save(&user).Error; err != nil {
		log.Printf("uploadAvatar: Ошибка сохранения аватара: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения аватара"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Аватар загружен", "profile": user})
}

func searchCourses(c *gin.Context) {
	query := c.Query("q")
	var courses []Course
	dbQuery := db.Preload("Teacher")
	if query != "" {
		jsonQuery := fmt.Sprintf(`["%s"]`, query)
		dbQuery = dbQuery.Where("title ILIKE ? OR description ILIKE ? OR services @> ?", "%"+query+"%", "%"+query+"%", jsonQuery)
	}
	if err := dbQuery.Find(&courses).Error; err != nil {
		log.Printf("searchCourses: Ошибка поиска курсов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка поиска"})
		return
	}
	log.Printf("Найдено %d курсов для запроса '%s'", len(courses), query)
	c.JSON(http.StatusOK, courses)
}

func getCourses(c *gin.Context) {
	userID := c.GetInt("userID")
	var courses []Course
	if err := db.Where("teacher_id = ?", userID).Find(&courses).Error; err != nil {
		log.Printf("getCourses: Ошибка получения курсов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения курсов"})
		return
	}
	c.JSON(http.StatusOK, courses)
}

func createCourse(c *gin.Context) {
	userID := c.GetInt("userID")
	role := c.GetString("role")
	if role != "nutri" {
		log.Println("createCourse: Доступ запрещён")
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ только для нутрициологов"})
		return
	}
	var input struct {
		Title       string      `json:"title"`
		Services    StringArray `json:"services"`
		Description string      `json:"description"`
		NetPrice    float64     `json:"net_price"`
		VideoURL    string      `json:"video_url"`
	}
	if err := c.BindJSON(&input); err != nil {
		log.Printf("createCourse: Неверные данные: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if input.NetPrice <= 0 {
		log.Println("createCourse: Неверная цена")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Чистая цена должна быть больше 0"})
		return
	}
	netPrice := decimal.NewFromFloat(input.NetPrice)
	grossPrice := netPrice.Mul(decimal.NewFromFloat(1.5))
	course := Course{
		TeacherID:   userID,
		Title:       input.Title,
		Services:    input.Services,
		Description: input.Description,
		NetPrice:    netPrice,
		GrossPrice:  grossPrice,
		VideoURL:    input.VideoURL,
	}
	if err := db.Create(&course).Error; err != nil {
		log.Printf("createCourse: Ошибка создания курса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания курса"})
		return
	}
	c.JSON(http.StatusOK, course)
}

func getCourse(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("getCourse: Неверный ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}
	var course Course
	if err := db.Preload("Teacher").First(&course, id).Error; err != nil {
		log.Printf("getCourse: Курс не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Курс не найден"})
		return
	}
	c.JSON(http.StatusOK, course)
}

func createPayment(c *gin.Context) {
	userID := c.GetInt("userID")
	role := c.GetString("role")
	if role != "client" {
		log.Println("createPayment: Доступ запрещён")
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ только для клиентов"})
		return
	}
	var input struct {
		CourseID int `json:"course_id"`
	}
	if err := c.BindJSON(&input); err != nil {
		log.Printf("createPayment: Неверные данные: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		log.Printf("createPayment: Пользователь не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}
	var course Course
	if err := db.First(&course, input.CourseID).Error; err != nil {
		log.Printf("createPayment: Курс не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Курс не найден"})
		return
	}
	var existingPayment Payment
	if err := db.Where("user_id = ? AND course_id = ? AND status = ?", userID, input.CourseID, "paid").First(&existingPayment).Error; err == nil {
		log.Printf("createPayment: Платеж за курс %d уже существует для пользователя %d", input.CourseID, userID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Вы уже оплатили этот курс"})
		return
	}
	grossAmount := course.GrossPrice
	commission := grossAmount.Sub(course.NetPrice)
	netAmount := course.NetPrice
	payment := Payment{
		UserID:      userID,
		CourseID:    input.CourseID,
		GrossAmount: grossAmount,
		Commission:  commission,
		NetAmount:   netAmount,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	if err := db.Create(&payment).Error; err != nil {
		log.Printf("createPayment: Ошибка создания платежа: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания платежа"})
		return
	}
	shopID := os.Getenv("SHOP_ID")
	secretKey := os.Getenv("SECRET_KEY")
	if shopID == "" || secretKey == "" {
		log.Println("createPayment: Отсутствуют SHOP_ID или SECRET_KEY")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка конфигурации платежной системы"})
		return
	}
	apiURL := "https://api.yookassa.ru/v3/payments"
	body := map[string]interface{}{
		"amount": map[string]interface{}{
			"value":    grossAmount.StringFixed(2),
			"currency": "RUB",
		},
		"confirmation": map[string]interface{}{
			"type":       "redirect",
			"return_url": "http://localhost:3000/return?payment_id=" + strconv.Itoa(payment.ID),
		},
		"capture":     true,
		"description": "Оплата услуги " + course.Title,
		"metadata": map[string]interface{}{
			"payment_id": payment.ID,
		},
		"receipt": map[string]interface{}{
			"customer": map[string]interface{}{
				"email": user.Email,
			},
			"items": []map[string]interface{}{
				{
					"description": course.Title,
					"quantity":    "1.00",
					"amount": map[string]interface{}{
						"value":    grossAmount.StringFixed(2),
						"currency": "RUB",
					},
					"vat_code": 1,
				},
			},
		},
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		log.Printf("createPayment: Ошибка сериализации тела запроса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания платежа"})
		return
	}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(bodyJSON))
	if err != nil {
		log.Printf("createPayment: Ошибка создания запроса к ЮKassa: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания платежа"})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", uuid.New().String())
	req.SetBasicAuth(shopID, secretKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("createPayment: Ошибка отправки запроса к ЮKassa: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка связи с платежной системой"})
		return
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("createPayment: Ошибка чтения ответа ЮKassa: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки ответа платежной системы"})
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("createPayment: Ошибка ЮKassa: status %d, body: %s", resp.StatusCode, string(bodyBytes))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания платежа в ЮKassa: " + string(bodyBytes)})
		return
	}
	var yookassaResp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &yookassaResp); err != nil {
		log.Printf("createPayment: Ошибка парсинга ответа ЮKassa: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки ответа платежной системы"})
		return
	}
	payment.YookassaID = yookassaResp["id"].(string)
	if err := db.Save(&payment).Error; err != nil {
		log.Printf("createPayment: Ошибка сохранения YookassaID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения данных платежа"})
		return
	}
	confirmation, ok := yookassaResp["confirmation"].(map[string]interface{})
	if !ok {
		log.Println("createPayment: Неверный формат confirmation в ответе ЮKassa")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки ответа платежной системы"})
		return
	}
	confirmationURL, ok := confirmation["confirmation_url"].(string)
	if !ok {
		log.Println("createPayment: Отсутствует confirmation_url в ответе ЮKassa")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки ответа платежной системы"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"confirmation_url": confirmationURL, "payment_id": payment.ID})
}

func returnPayment(c *gin.Context) {
	paymentIDStr := c.Query("payment_id")
	paymentID, err := strconv.Atoi(paymentIDStr)
	if err != nil {
		log.Printf("returnPayment: Неверный ID платежа: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID платежа"})
		return
	}
	var payment Payment
	if err := db.First(&payment, paymentID).Error; err != nil {
		log.Printf("returnPayment: Платеж не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Платеж не найден"})
		return
	}
	if payment.YookassaID == "" {
		log.Println("returnPayment: Платеж не инициализирован")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Платеж не инициализирован"})
		return
	}
	shopID := os.Getenv("SHOP_ID")
	secretKey := os.Getenv("SECRET_KEY")
	apiURL := "https://api.yookassa.ru/v3/payments/" + payment.YookassaID
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("returnPayment: Ошибка создания запроса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки платежа"})
		return
	}
	req.SetBasicAuth(shopID, secretKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("returnPayment: Ошибка запроса к ЮKassa: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки платежа"})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyErr, _ := io.ReadAll(resp.Body)
		log.Printf("returnPayment: Ошибка ЮKassa: status %d, body: %s", resp.StatusCode, string(bodyErr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки платежа в ЮKassa"})
		return
	}
	var yookassaResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&yookassaResp); err != nil {
		log.Printf("returnPayment: Ошибка парсинга ответа: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки платежа"})
		return
	}
	status := yookassaResp["status"].(string)
	if status == "succeeded" && payment.Status != "paid" {
		payment.Status = "paid"
		payment.TransactionID = yookassaResp["id"].(string)
		if err := db.Save(&payment).Error; err != nil {
			log.Printf("returnPayment: Ошибка обновления платежа: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления платежа"})
			return
		}
		var course Course
		if err := db.First(&course, payment.CourseID).Error; err != nil {
			log.Printf("returnPayment: Курс не найден: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Курс не найден"})
			return
		}
		if err := db.Model(&User{}).Where("id = ?", course.TeacherID).Update("balance", gorm.Expr("balance + ?", payment.NetAmount)).Error; err != nil {
			log.Printf("returnPayment: Ошибка начисления баланса: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка начисления баланса"})
			return
		}
		enrollment := Enrollment{
			CourseID: payment.CourseID,
			UserID:   payment.UserID,
		}
		if err := db.Create(&enrollment).Error; err != nil {
			log.Printf("returnPayment: Ошибка создания записи: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания записи"})
			return
		}
	} else if status == "pending" {
		c.JSON(http.StatusOK, gin.H{"status": "pending", "message": "Оплата в обработке"})
		return
	} else if status == "canceled" || status == "failed" {
		payment.Status = "failed"
		if err := db.Save(&payment).Error; err != nil {
			log.Printf("returnPayment: Ошибка обновления платежа: %v", err)
		}
		c.JSON(http.StatusOK, gin.H{"status": "failed", "message": "Оплата не удалась"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": payment.Status, "message": "Оплата успешна", "transaction_id": payment.TransactionID})
}

func webhookYookassa(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("webhookYookassa: Ошибка чтения webhook: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}
	signatureHeader := c.GetHeader("Content-Signature")
	if signatureHeader == "" {
		log.Println("webhookYookassa: Signature отсутствует")
		c.Status(http.StatusBadRequest)
		return
	}
	secretKey := []byte(os.Getenv("SECRET_KEY"))
	h := hmac.New(sha256.New, secretKey)
	h.Write(body)
	expectedSignature := "sha256=" + hex.EncodeToString(h.Sum(nil))
	if signatureHeader != expectedSignature {
		log.Printf("webhookYookassa: Неверная signature: expected %s, got %s", expectedSignature, signatureHeader)
		c.Status(http.StatusBadRequest)
		return
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("webhookYookassa: Ошибка парсинга webhook: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}
	event := payload["event"].(string)
	if event == "payment.succeeded" {
		object := payload["object"].(map[string]interface{})
		yookassaID := object["id"].(string)
		var payment Payment
		if err := db.Where("yookassa_id = ?", yookassaID).First(&payment).Error; err != nil {
			log.Printf("webhookYookassa: Платеж не найден: %v", err)
			c.Status(http.StatusOK)
			return
		}
		if payment.Status == "paid" {
			c.Status(http.StatusOK)
			return
		}
		payment.Status = "paid"
		payment.TransactionID = yookassaID
		if err := db.Save(&payment).Error; err != nil {
			log.Printf("webhookYookassa: Ошибка обновления платежа: %v", err)
			c.Status(http.StatusOK)
			return
		}
		var course Course
		if err := db.First(&course, payment.CourseID).Error; err != nil {
			log.Printf("webhookYookassa: Курс не найден: %v", err)
			c.Status(http.StatusOK)
			return
		}
		if err := db.Model(&User{}).Where("id = ?", course.TeacherID).Update("balance", gorm.Expr("balance + ?", payment.NetAmount)).Error; err != nil {
			log.Printf("webhookYookassa: Ошибка начисления баланса: %v", err)
			c.Status(http.StatusOK)
			return
		}
		enrollment := Enrollment{
			CourseID: payment.CourseID,
			UserID:   payment.UserID,
		}
		if err := db.Create(&enrollment).Error; err != nil {
			log.Printf("webhookYookassa: Ошибка создания записи: %v", err)
			c.Status(http.StatusOK)
			return
		}
		notification := Notification{
			UserID:  course.TeacherID,
			Type:    "payment",
			Content: "Получена оплата за услугу " + course.Title + ": " + payment.NetAmount.StringFixed(2) + " руб.",
		}
		if err := db.Create(&notification).Error; err != nil {
			log.Printf("webhookYookassa: Ошибка создания уведомления: %v", err)
		}
		notifJSON, _ := json.Marshal(map[string]interface{}{
			"type": "notification",
			"data": notification,
		})
		sendToUser(course.TeacherID, notifJSON)
	}
	c.Status(http.StatusOK)
}

func getUserReviews(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("getUserReviews: Неверный ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}
	var reviews []Review
	if err := db.Preload("Author").Where("course_id IN (SELECT id FROM courses WHERE teacher_id = ?)", id).Find(&reviews).Error; err != nil {
		log.Printf("getUserReviews: Ошибка получения отзывов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения отзывов"})
		return
	}
	c.JSON(http.StatusOK, reviews)
}

func getCourseReviews(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("getCourseReviews: Неверный ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}
	var reviews []Review
	if err := db.Preload("Author").Where("course_id = ?", id).Find(&reviews).Error; err != nil {
		log.Printf("getCourseReviews: Ошибка получения отзывов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения отзывов"})
		return
	}
	c.JSON(http.StatusOK, reviews)
}

func getRandomReviews(c *gin.Context) {
	var reviews []Review
	if err := db.Preload("Author").Order("RANDOM()").Limit(6).Find(&reviews).Error; err != nil {
		log.Printf("getRandomReviews: Ошибка получения отзывов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения отзывов"})
		return
	}
	c.JSON(http.StatusOK, reviews)
}

func createReview(c *gin.Context) {
	userID := c.GetInt("userID")
	var input struct {
		CourseID int    `json:"course_id"`
		Content  string `json:"content"`
	}
	if err := c.BindJSON(&input); err != nil {
		log.Printf("createReview: Неверные данные: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	var enrollment Enrollment
	if err := db.Where("user_id = ? AND course_id = ?", userID, input.CourseID).First(&enrollment).Error; err != nil {
		log.Printf("createReview: Отзыв запрещён: %v", err)
		c.JSON(http.StatusForbidden, gin.H{"error": "Отзыв возможен только после оплаты"})
		return
	}
	review := Review{
		AuthorID: userID,
		CourseID: input.CourseID,
		Content:  input.Content,
	}
	if err := db.Create(&review).Error; err != nil {
		log.Printf("createReview: Ошибка создания отзыва: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания отзыва"})
		return
	}
	c.JSON(http.StatusOK, review)
}

func getEnrolled(c *gin.Context) {
	userID := c.GetInt("userID")
	var enrollments []Enrollment
	if err := db.Preload("Course.Teacher").Where("user_id = ?", userID).Find(&enrollments).Error; err != nil {
		log.Printf("getEnrolled: Ошибка получения записей: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения записей"})
		return
	}
	c.JSON(http.StatusOK, enrollments)
}

func getNutris(c *gin.Context) {
	limitStr := c.Query("limit")
	random := c.Query("random") == "true"
	var nutris []User
	dbQuery := db.Where("role = ?", "nutri")
	if random {
		dbQuery = dbQuery.Order("RANDOM()")
	}
	if limitStr != "" {
		limit, _ := strconv.Atoi(limitStr)
		dbQuery = dbQuery.Limit(limit)
	}
	if err := dbQuery.Find(&nutris).Error; err != nil {
		log.Printf("getNutris: Ошибка получения нутрициологов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения нутрициологов"})
		return
	}
	log.Printf("Возвращено %d нутрициологов", len(nutris))
	c.JSON(http.StatusOK, nutris)
}

func getAdminNutris(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		log.Println("getAdminNutris: Доступ запрещён")
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ только для администраторов"})
		return
	}
	var nutris []User
	if err := db.Where("role = ?", "nutri").Find(&nutris).Error; err != nil {
		log.Printf("getAdminNutris: Ошибка получения нутрициологов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения нутрициологов"})
		return
	}
	log.Printf("Возвращено %d нутрициологов для администратора", len(nutris))
	c.JSON(http.StatusOK, nutris)
}

func decryptCard(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		log.Println("decryptCard: Доступ запрещён")
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ только для администраторов"})
		return
	}
	var input struct {
		UserID int `json:"user_id"`
	}
	if err := c.BindJSON(&input); err != nil {
		log.Printf("decryptCard: Неверные данные: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	var user User
	if err := db.First(&user, input.UserID).Error; err != nil {
		log.Printf("decryptCard: Пользователь не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}
	if user.EncryptedCard == "" {
		log.Println("decryptCard: Карта не указана")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Карта не указана"})
		return
	}
	key, err := hex.DecodeString(os.Getenv("AES_KEY"))
	if err != nil {
		log.Printf("decryptCard: Ошибка декодирования ключа AES: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка расшифровки"})
		return
	}
	ciphertext, err := hex.DecodeString(user.EncryptedCard)
	if err != nil {
		log.Printf("decryptCard: Ошибка декодирования карты: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка расшифровки"})
		return
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("decryptCard: Ошибка создания шифра: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка расшифровки"})
		return
	}
	if len(ciphertext) < aes.BlockSize {
		log.Println("decryptCard: Неверный размер ciphertext")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка расшифровки"})
		return
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	if len(ciphertext)%aes.BlockSize != 0 {
		log.Println("decryptCard: Неверный размер ciphertext")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка расшифровки"})
		return
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)
	plaintext = bytes.TrimRight(plaintext, "\x00")
	c.JSON(http.StatusOK, gin.H{"card_number": string(plaintext)})
}

func processPayout(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		log.Println("processPayout: Доступ запрещён")
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ только для администраторов"})
		return
	}
	var input struct {
		UserID int             `json:"user_id"`
		Amount decimal.Decimal `json:"amount"`
	}
	if err := c.BindJSON(&input); err != nil {
		log.Printf("processPayout: Неверные данные: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	var user User
	if err := db.First(&user, input.UserID).Error; err != nil {
		log.Printf("processPayout: Пользователь не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}
	if input.Amount.LessThanOrEqual(decimal.Zero) {
		log.Println("processPayout: Сумма должна быть больше 0")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Сумма должна быть больше 0"})
		return
	}
	if input.Amount.GreaterThan(user.Balance) {
		log.Println("processPayout: Недостаточно средств на балансе")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недостаточно средств на балансе"})
		return
	}
	if err := db.Model(&user).Update("balance", gorm.Expr("balance - ?", input.Amount)).Error; err != nil {
		log.Printf("processPayout: Ошибка списания баланса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка списания баланса"})
		return
	}
	if err := db.Model(&user).Update("payout_amount", gorm.Expr("payout_amount + ?", input.Amount)).Error; err != nil {
		log.Printf("processPayout: Ошибка обновления выплаченной суммы: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления выплаченной суммы"})
		return
	}
	log.Printf("Инициирована выплата %s руб. для пользователя %d", input.Amount.StringFixed(2), input.UserID)
	c.JSON(http.StatusOK, gin.H{"message": "Выплата инициирована"})
}

func updatePayoutAmount(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" {
		log.Println("updatePayoutAmount: Доступ запрещён")
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ только для администраторов"})
		return
	}
	var input struct {
		UserID       int             `json:"user_id"`
		PayoutAmount decimal.Decimal `json:"payout_amount"`
	}
	if err := c.BindJSON(&input); err != nil {
		log.Printf("updatePayoutAmount: Неверные данные: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	var user User
	if err := db.First(&user, input.UserID).Error; err != nil {
		log.Printf("updatePayoutAmount: Пользователь не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}
	if input.PayoutAmount.LessThan(decimal.Zero) {
		log.Println("updatePayoutAmount: Выплаченная сумма не может быть отрицательной")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Выплаченная сумма не может быть отрицательной"})
		return
	}
	if err := db.Model(&user).Update("payout_amount", input.PayoutAmount).Error; err != nil {
		log.Printf("updatePayoutAmount: Ошибка обновления выплаченной суммы: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления выплаченной суммы"})
		return
	}
	log.Printf("Обновлена выплаченная сумма %s руб. для пользователя %d", input.PayoutAmount.StringFixed(2), input.UserID)
	c.JSON(http.StatusOK, gin.H{"message": "Выплаченная сумма обновлена"})
}

func startChat(c *gin.Context) {
	userID := c.GetInt("userID")
	var input struct {
		ReceiverID int `json:"receiver_id"`
	}
	if err := c.BindJSON(&input); err != nil {
		log.Printf("startChat: Неверные данные: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if userID == input.ReceiverID {
		log.Println("startChat: Нельзя начать чат с самим собой")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя начать чат с самим собой"})
		return
	}
	var count int64
	if err := db.Model(&Dialog{}).Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, input.ReceiverID, input.ReceiverID, userID).Count(&count).Error; err != nil {
		log.Printf("startChat: Ошибка проверки диалога: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка начала чата"})
		return
	}
	if count == 0 {
		dialog := Dialog{
			SenderID:   userID,
			ReceiverID: input.ReceiverID,
		}
		if err := db.Create(&dialog).Error; err != nil {
			log.Printf("startChat: Ошибка создания диалога: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка начала чата"})
			return
		}
		log.Printf("Создан диалог %d -> %d", userID, input.ReceiverID)
	}
	chatStartedJSON, err := json.Marshal(map[string]interface{}{
		"type": "chat:started",
		"data": map[string]int{"receiver_id": input.ReceiverID},
	})
	if err != nil {
		log.Printf("startChat: Ошибка marshal chat:started: %v", err)
	} else {
		sendToUser(userID, chatStartedJSON)
	}
	c.JSON(http.StatusOK, gin.H{"receiver_id": input.ReceiverID})
}

func getChats(c *gin.Context) {
	userID := c.GetInt("userID")
	var dialogs []struct {
		UserID      int    `json:"user_id"`
		FullName    string `json:"full_name"`
		AvatarURL   string `json:"avatar_url"`
		LastMessage string `json:"last_message"`
		UnreadCount int    `json:"unread_count"`
	}
	err := db.Raw(`
		SELECT u.id as user_id, u.full_name, u.avatar_url, 
		       (SELECT content FROM messages WHERE (sender_id = u.id AND receiver_id = ? OR sender_id = ? AND receiver_id = u.id) ORDER BY created_at DESC LIMIT 1) as last_message,
		       (SELECT COUNT(*) FROM messages WHERE sender_id = u.id AND receiver_id = ? AND read_at IS NULL) as unread_count
		FROM users u
		WHERE EXISTS (SELECT 1 FROM dialogs WHERE (sender_id = u.id AND receiver_id = ?) OR (sender_id = ? AND receiver_id = u.id))
		AND u.id != ?
		ORDER BY (
			SELECT COALESCE(
				(SELECT MAX(m.created_at) FROM messages m WHERE (m.sender_id = u.id AND m.receiver_id = ?) OR (m.sender_id = ? AND m.receiver_id = u.id)),
				(SELECT d.created_at FROM dialogs d WHERE (d.sender_id = u.id AND d.receiver_id = ?) OR (d.sender_id = ? AND d.receiver_id = u.id) LIMIT 1)
			)
		) DESC
	`, userID, userID, userID, userID, userID, userID, userID, userID, userID, userID).Scan(&dialogs).Error
	if err != nil {
		log.Printf("getChats: Ошибка получения чатов: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения чатов"})
		return
	}
	log.Printf("Загружено %d диалогов для пользователя %d", len(dialogs), userID)
	c.JSON(http.StatusOK, dialogs)
}

func getMessages(c *gin.Context) {
	userID := c.GetInt("userID")
	receiverIDStr := c.Query("receiver_id")
	receiverID, err := strconv.Atoi(receiverIDStr)
	if err != nil {
		log.Printf("getMessages: Неверный ID получателя: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID получателя"})
		return
	}
	var messages []Message
	if err := db.Preload("Sender").Preload("Receiver").
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, receiverID, receiverID, userID).
		Order("created_at ASC").Find(&messages).Error; err != nil {
		log.Printf("getMessages: Ошибка получения сообщений: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения сообщений"})
		return
	}
	c.JSON(http.StatusOK, messages)
}

func sendMessage(c *gin.Context) {
	userID := c.GetInt("userID")
	var input struct {
		ReceiverID int    `json:"receiver_id"`
		Content    string `json:"content"`
	}
	if err := c.BindJSON(&input); err != nil {
		log.Printf("sendMessage: Неверные данные: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	message := Message{
		SenderID:   userID,
		ReceiverID: input.ReceiverID,
		Content:    input.Content,
	}
	if err := db.Create(&message).Error; err != nil {
		log.Printf("sendMessage: Ошибка отправки сообщения: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отправки сообщения"})
		return
	}
	if err := db.First(&message).Error; err != nil {
		log.Printf("sendMessage: Ошибка загрузки сообщения: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки сообщения"})
		return
	}
	msgJSON, err := json.Marshal(map[string]interface{}{
		"type": "message",
		"data": message,
	})
	if err != nil {
		log.Printf("sendMessage: Ошибка marshal сообщения: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки сообщения"})
		return
	}
	if conn, ok := clients[userID]; ok {
		if err := conn.WriteMessage(websocket.TextMessage, msgJSON); err != nil {
			log.Printf("sendMessage: Ошибка отправки WebSocket сообщения пользователю %d: %v", userID, err)
			conn.Close()
			delete(clients, userID)
		}
	}
	if conn, ok := clients[input.ReceiverID]; ok {
		if err := conn.WriteMessage(websocket.TextMessage, msgJSON); err != nil {
			log.Printf("sendMessage: Ошибка отправки WebSocket сообщения пользователю %d: %v", input.ReceiverID, err)
			conn.Close()
			delete(clients, input.ReceiverID)
		}
	}
	log.Printf("Отправлено сообщение от %d к %d: %s", userID, input.ReceiverID, input.Content)
	c.JSON(http.StatusOK, message)
}

func markRead(c *gin.Context) {
	userID := c.GetInt("userID")
	var input struct {
		ReceiverID int `json:"receiver_id"`
	}
	if err := c.BindJSON(&input); err != nil {
		log.Printf("markRead: Неверные данные: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if err := db.Model(&Message{}).Where("sender_id = ? AND receiver_id = ? AND read_at IS NULL", input.ReceiverID, userID).
		Update("read_at", time.Now()).Error; err != nil {
		log.Printf("markRead: Ошибка отметки прочитанных: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отметки прочитанных"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Сообщения отмечены прочитанными"})
}

func handleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("handleWebSocket: Ошибка апгрейда WebSocket: %v", err)
		return
	}
	tokenString := c.Request.URL.Query().Get("token")
	if tokenString == "" {
		log.Println("handleWebSocket: Токен не предоставлен")
		conn.WriteMessage(websocket.TextMessage, []byte("Токен не предоставлен"))
		conn.Close()
		return
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		log.Printf("handleWebSocket: Неверный токен: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Неверный токен"))
		conn.Close()
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("handleWebSocket: Неверные данные токена")
		conn.WriteMessage(websocket.TextMessage, []byte("Неверные данные токена"))
		conn.Close()
		return
	}
	userID := int(claims["id"].(float64))
	log.Printf("WebSocket подключен для пользователя %d", userID)
	if oldConn, exists := clients[userID]; exists {
		oldConn.Close()
		delete(clients, userID)
	}
	clients[userID] = conn
	defer func() {
		log.Printf("WebSocket отключен для пользователя %d", userID)
		conn.Close()
		delete(clients, userID)
	}()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("handleWebSocket: Ошибка чтения WebSocket для пользователя %d: %v", userID, err)
			break
		}
	}
}

func sendToUser(userID int, message []byte) {
	if conn, ok := clients[userID]; ok {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("sendToUser: Ошибка отправки WebSocket сообщения пользователю %d: %v", userID, err)
			conn.Close()
			delete(clients, userID)
		}
	} else {
		log.Printf("sendToUser: Клиент %d не подключен", userID)
	}
}
