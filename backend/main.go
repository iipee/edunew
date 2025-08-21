package main

import (
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

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

type Review struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	CourseID  int       `json:"course_id"`
	AuthorID  int       `json:"author_id"`
	Content   string    `json:"content" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	Author    User      `json:"author" gorm:"foreignKey:AuthorID"`
}

type Enrollment struct {
	ID        int       `json:"id" gorm:"primaryKey"`
	UserID    int       `json:"user_id"`
	CourseID  int       `json:"course_id"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
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
	UpdatedAt     time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

type Message struct {
	ID         int        `json:"id" gorm:"primaryKey"`
	SenderID   int        `json:"sender_id"`
	ReceiverID int        `json:"receiver_id"`
	Content    string     `json:"content" gorm:"not null"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	ReadAt     *time.Time `json:"read_at"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Ошибка загрузки .env файла:", err)
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN не установлен в .env")
	}

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	err = db.AutoMigrate(&User{}, &Course{}, &Review{}, &Enrollment{}, &Message{}, &Payment{})
	if err != nil {
		log.Fatalf("Ошибка миграции БД: %v", err)
	}

	r = gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.POST("/api/register", register)
	r.POST("/api/login", login)
	r.GET("/api/profile", authMiddleware, getProfile)
	r.PUT("/api/profile", authMiddleware, updateProfile)
	r.POST("/api/profile/update-card", authMiddleware, handleUpdateCard)
	r.GET("/api/profile/:id", getOtherProfile)
	r.GET("/api/nutris", getNutris)
	r.GET("/api/search", searchCourses)
	r.POST("/api/courses", authMiddleware, createCourse)
	r.GET("/api/courses/:id", getCourse)
	r.POST("/api/reviews", authMiddleware, createReview)
	r.GET("/api/reviews/course/:id", getCourseReviews)
	r.GET("/api/reviews/user/:id", getUserReviews)
	r.GET("/api/reviews/random", getRandomReviews)
	r.GET("/api/enrolled", authMiddleware, getEnrolled)
	r.POST("/api/payments/create", authMiddleware, createPayment)
	r.GET("/api/payments/return", authMiddleware, getPaymentReturn)
	r.POST("/webhook/yookassa", webhookHandler)
	r.POST("/api/start-chat", authMiddleware, startChat)
	r.GET("/api/chats", authMiddleware, getChats)
	r.GET("/api/messages", authMiddleware, getMessages)
	r.POST("/api/messages", authMiddleware, sendMessage)
	r.PUT("/api/messages/read", authMiddleware, markRead)
	r.GET("/ws", func(c *gin.Context) {
		handleWebSocket(c.Writer, c.Request)
	})

	log.Fatal(r.Run(":8080"))
}

func authMiddleware(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Токен не предоставлен"})
		c.Abort()
		return
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
		c.Abort()
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные данные токена"})
		c.Abort()
		return
	}
	userID := int(claims["id"].(float64))
	c.Set("userID", userID)
	c.Set("role", claims["role"])
	c.Next()
}

func register(c *gin.Context) {
	var input struct {
		FullName    string `json:"full_name"`
		Username    string `json:"username"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		Role        string `json:"role"`
		Description string `json:"description"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 14)
	if err != nil {
		log.Printf("Ошибка хеширования пароля: %v", err)
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
		log.Printf("Ошибка регистрации: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка регистрации"})
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":   user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		log.Printf("Ошибка создания токена: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания токена"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	var user User
	if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные данные"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
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
		log.Printf("Ошибка создания токена: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания токена"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString, "role": user.Role, "id": user.ID})
}

func getProfile(c *gin.Context) {
	userID := c.GetInt("userID")
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		log.Printf("Пользователь не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}
	var courses []Course
	db.Where("teacher_id = ?", userID).Find(&courses)
	c.JSON(http.StatusOK, gin.H{"profile": user, "courses": courses})
}

func updateProfile(c *gin.Context) {
	userID := c.GetInt("userID")
	var input struct {
		FullName    string `json:"full_name"`
		Description string `json:"description"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if err := db.Model(&User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"full_name":   input.FullName,
		"description": input.Description,
	}).Error; err != nil {
		log.Printf("Ошибка обновления профиля: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Профиль обновлен"})
}

func handleUpdateCard(c *gin.Context) {
	userID := c.GetInt("userID")
	var input struct {
		CardNumber string `json:"card_number"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if len(input.CardNumber) != 16 || !regexp.MustCompile(`^\d{16}$`).MatchString(input.CardNumber) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный номер карты (16 цифр)"})
		return
	}
	key := []byte(os.Getenv("AES_KEY"))
	if len(key) != 32 {
		log.Printf("Ошибка: AES_KEY должен быть 32 байта")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка конфигурации шифрования"})
		return
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Printf("Ошибка создания AES: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}
	plaintext := []byte(input.CardNumber)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Printf("Ошибка IV: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
		return
	}
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
	encrypted := hex.EncodeToString(ciphertext)
	if err := db.Model(&User{}).Where("id = ?", userID).Update("encrypted_card", encrypted).Error; err != nil {
		log.Printf("Ошибка обновления карты: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Карта обновлена"})
}

func getOtherProfile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID пользователя"})
		return
	}
	var user User
	if err := db.First(&user, id).Error; err != nil {
		log.Printf("Пользователь не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}
	var courses []Course
	db.Where("teacher_id = ?", id).Find(&courses)
	c.JSON(http.StatusOK, gin.H{"profile": user, "courses": courses})
}

func getNutris(c *gin.Context) {
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = 10
	}
	random := c.Query("random") == "true"
	var users []User
	query := db.Where("role = ?", "nutri").Limit(limit)
	if random {
		query = query.Order("RANDOM()")
	}
	query.Find(&users)
	c.JSON(http.StatusOK, users)
}

func searchCourses(c *gin.Context) {
	query := c.Query("q")
	var courses []Course
	if query == "" {
		db.Preload("Teacher").Find(&courses)
	} else {
		db.Preload("Teacher").Where("title ILIKE ? OR description ILIKE ? OR ? = ANY(services)", "%"+query+"%", "%"+query+"%", query).Find(&courses)
	}
	c.JSON(http.StatusOK, courses)
}

func createCourse(c *gin.Context) {
	userID := c.GetInt("userID")
	role := c.GetString("role")
	if role != "nutri" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ только для нутрициологов"})
		return
	}
	var input struct {
		Title       string   `json:"title"`
		Services    []string `json:"services"`
		Description string   `json:"description"`
		NetPrice    float64  `json:"net_price"`
		VideoURL    string   `json:"video_url"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if input.NetPrice <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Цена должна быть больше 0"})
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
		log.Printf("Ошибка создания курса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания курса"})
		return
	}
	c.JSON(http.StatusOK, course)
}

func getCourse(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID курса"})
		return
	}
	var course Course
	if err := db.Preload("Teacher").First(&course, id).Error; err != nil {
		log.Printf("Курс не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Курс не найден"})
		return
	}
	c.JSON(http.StatusOK, course)
}

func createReview(c *gin.Context) {
	userID := c.GetInt("userID")
	var input struct {
		CourseID int    `json:"course_id"`
		Content  string `json:"content"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	var enrollment Enrollment
	if err := db.Where("user_id = ? AND course_id = ?", userID, input.CourseID).First(&enrollment).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Отзыв только после оплаты"})
		return
	}
	review := Review{
		CourseID: input.CourseID,
		AuthorID: userID,
		Content:  input.Content,
	}
	if err := db.Create(&review).Error; err != nil {
		log.Printf("Ошибка создания отзыва: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания отзыва"})
		return
	}
	c.JSON(http.StatusOK, review)
}

func getCourseReviews(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID курса"})
		return
	}
	var reviews []Review
	db.Preload("Author").Where("course_id = ?", id).Find(&reviews)
	c.JSON(http.StatusOK, reviews)
}

func getUserReviews(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID пользователя"})
		return
	}
	var reviews []Review
	db.Preload("Author").Where("author_id = ?", id).Find(&reviews)
	c.JSON(http.StatusOK, reviews)
}

func getRandomReviews(c *gin.Context) {
	var reviews []Review
	db.Preload("Author").Order("RANDOM()").Limit(3).Find(&reviews)
	c.JSON(http.StatusOK, reviews)
}

func getEnrolled(c *gin.Context) {
	userID := c.GetInt("userID")
	var enrollments []Enrollment
	db.Where("user_id = ?", userID).Find(&enrollments)
	var response []struct {
		ID        int       `json:"id"`
		UserID    int       `json:"user_id"`
		CourseID  int       `json:"course_id"`
		CreatedAt time.Time `json:"created_at"`
		Course    Course    `json:"course"`
	}
	for _, e := range enrollments {
		var course Course
		if err := db.Preload("Teacher").First(&course, e.CourseID).Error; err != nil {
			log.Printf("Курс не найден для enrollment %d: %v", e.ID, err)
			continue
		}
		response = append(response, struct {
			ID        int       `json:"id"`
			UserID    int       `json:"user_id"`
			CourseID  int       `json:"course_id"`
			CreatedAt time.Time `json:"created_at"`
			Course    Course    `json:"course"`
		}{
			ID:        e.ID,
			UserID:    e.UserID,
			CourseID:  e.CourseID,
			CreatedAt: e.CreatedAt,
			Course:    course,
		})
	}
	c.JSON(http.StatusOK, response)
}

func createPayment(c *gin.Context) {
	userID := c.GetInt("userID")
	var input struct {
		CourseID int `json:"course_id"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	var course Course
	if err := db.First(&course, input.CourseID).Error; err != nil {
		log.Printf("Курс не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Курс не найден"})
		return
	}
	netAmount := course.NetPrice
	grossAmount := course.GrossPrice
	commission := grossAmount.Sub(netAmount)

	shopID := os.Getenv("SHOP_ID")
	secretKey := os.Getenv("SECRET_KEY")
	if shopID == "" || secretKey == "" {
		log.Printf("Ошибка: SHOP_ID или SECRET_KEY не установлены")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка конфигурации платежей"})
		return
	}

	yookassaID, confirmationURL, err := createYookassaPayment(grossAmount.StringFixed(2))
	if err != nil {
		log.Printf("Ошибка создания платежа ЮKassa: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания платежа"})
		return
	}

	payment := Payment{
		UserID:        userID,
		CourseID:      input.CourseID,
		GrossAmount:   grossAmount,
		Commission:    commission,
		NetAmount:     netAmount,
		Status:        "pending",
		YookassaID:    yookassaID,
		TransactionID: yookassaID,
	}
	if err := db.Create(&payment).Error; err != nil {
		log.Printf("Ошибка сохранения платежа: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения платежа"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"confirmation_url": confirmationURL, "payment_id": payment.ID})
}

func createYookassaPayment(amount string) (string, string, error) {
	shopID := os.Getenv("SHOP_ID")
	secretKey := os.Getenv("SECRET_KEY")
	returnURL := "http://localhost:3000/return"

	body := map[string]interface{}{
		"amount": map[string]string{
			"value":    amount,
			"currency": "RUB",
		},
		"confirmation": map[string]string{
			"type":       "redirect",
			"return_url": returnURL,
		},
		"capture":     true,
		"description": "Оплата курса",
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return "", "", fmt.Errorf("ошибка сериализации JSON: %v", err)
	}
	req, err := http.NewRequest("POST", "https://sandbox.yookassa.ru/v3/payments", strings.NewReader(string(jsonBody)))
	if err != nil {
		return "", "", fmt.Errorf("ошибка создания запроса: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", fmt.Sprintf("%d", time.Now().UnixNano()))
	req.SetBasicAuth(shopID, secretKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("ошибка выполнения запроса: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", fmt.Errorf("ошибка декодирования ответа: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("ошибка ЮKassa: %v", result["error"])
	}
	yookassaID, ok := result["id"].(string)
	if !ok {
		return "", "", fmt.Errorf("ошибка: неверный формат ID платежа")
	}
	confirmation, ok := result["confirmation"].(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("ошибка: неверный формат confirmation")
	}
	confirmationURL, ok := confirmation["confirmation_url"].(string)
	if !ok {
		return "", "", fmt.Errorf("ошибка: неверный формат confirmation_url")
	}
	return yookassaID, confirmationURL, nil
}

func getPaymentReturn(c *gin.Context) {
	paymentID, err := strconv.Atoi(c.Query("payment_id"))
	if err != nil {
		log.Printf("Неверный ID платежа: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID платежа"})
		return
	}
	var payment Payment
	if err := db.First(&payment, paymentID).Error; err != nil {
		log.Printf("Платеж не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Платеж не найден"})
		return
	}
	message := "Ошибка оплаты"
	if payment.Status == "paid" {
		message = "Оплата успешна"
	}
	c.JSON(http.StatusOK, gin.H{
		"status":         payment.Status,
		"message":        message,
		"transaction_id": payment.TransactionID,
	})
}

func webhookHandler(c *gin.Context) {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		log.Printf("Ошибка: SECRET_KEY не установлен")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка конфигурации"})
		return
	}
	signature := c.GetHeader("Content-Signature")
	if signature == "" {
		log.Printf("Ошибка: Подпись не предоставлена")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Подпись не предоставлена"})
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела запроса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения тела запроса"})
		return
	}
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write(body)
	expected := "sha256=" + hex.EncodeToString(h.Sum(nil))
	if signature != expected {
		log.Printf("Неверная подпись webhook: %s != %s", signature, expected)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверная подпись"})
		return
	}
	var notification map[string]interface{}
	if err := json.Unmarshal(body, &notification); err != nil {
		log.Printf("Ошибка разбора уведомления: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат уведомления"})
		return
	}
	event, ok := notification["event"].(string)
	if !ok {
		log.Printf("Ошибка: Неверный формат события")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат события"})
		return
	}
	if event == "payment.succeeded" {
		object, ok := notification["object"].(map[string]interface{})
		if !ok {
			log.Printf("Ошибка: Неверный формат объекта")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат объекта"})
			return
		}
		yookassaID, ok := object["id"].(string)
		if !ok {
			log.Printf("Ошибка: Неверный формат ID платежа")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID платежа"})
			return
		}
		var payment Payment
		if err := db.Where("yookassa_id = ?", yookassaID).First(&payment).Error; err != nil {
			log.Printf("Платеж не найден: %s", yookassaID)
			c.Status(http.StatusOK)
			return
		}
		payment.Status = "paid"
		payment.TransactionID = yookassaID
		if err := db.Save(&payment).Error; err != nil {
			log.Printf("Ошибка обновления статуса: %v", err)
			c.Status(http.StatusOK)
			return
		}
		var course Course
		if err := db.First(&course, payment.CourseID).Error; err != nil {
			log.Printf("Курс не найден: %d", payment.CourseID)
			c.Status(http.StatusOK)
			return
		}
		if err := db.Model(&User{}).Where("id = ?", course.TeacherID).Update("balance", gorm.Expr("balance + ?", payment.NetAmount)).Error; err != nil {
			log.Printf("Ошибка обновления баланса: %v", err)
			c.Status(http.StatusOK)
			return
		}
		enrollment := Enrollment{UserID: payment.UserID, CourseID: payment.CourseID}
		if err := db.Create(&enrollment).Error; err != nil {
			log.Printf("Ошибка создания записи о зачислении: %v", err)
			c.Status(http.StatusOK)
			return
		}
	}
	c.Status(http.StatusOK)
}

func startChat(c *gin.Context) {
	var input struct {
		ReceiverID int `json:"receiver_id"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	var receiver User
	if err := db.First(&receiver, input.ReceiverID).Error; err != nil {
		log.Printf("Получатель не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Получатель не найден"})
		return
	}
	if input.ReceiverID == c.GetInt("userID") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Нельзя начать чат с самим собой"})
		return
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
	db.Raw(`
		SELECT DISTINCT ON (u.id) u.id AS user_id, u.full_name, u.avatar_url, m.content AS last_message,
		COUNT(CASE WHEN m.read_at IS NULL AND m.receiver_id = ? THEN 1 END) AS unread_count
		FROM users u
		JOIN messages m ON (m.sender_id = u.id AND m.receiver_id = ?) OR (m.sender_id = ? AND m.receiver_id = u.id)
		WHERE u.id != ?
		GROUP BY u.id, m.content
		ORDER BY u.id, m.created_at DESC
	`, userID, userID, userID, userID).Scan(&dialogs)
	c.JSON(http.StatusOK, dialogs)
}

func getMessages(c *gin.Context) {
	userID := c.GetInt("userID")
	receiverID, err := strconv.Atoi(c.Query("receiver_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID получателя"})
		return
	}
	var messages []Message
	db.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)", userID, receiverID, receiverID, userID).
		Order("created_at ASC").Find(&messages)
	c.JSON(http.StatusOK, messages)
}

func sendMessage(c *gin.Context) {
	userID := c.GetInt("userID")
	var input struct {
		ReceiverID int    `json:"receiver_id"`
		Content    string `json:"content"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if len(input.Content) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Сообщение слишком длинное"})
		return
	}
	var receiver User
	if err := db.First(&receiver, input.ReceiverID).Error; err != nil {
		log.Printf("Получатель не найден: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Получатель не найден"})
		return
	}
	message := Message{
		SenderID:   userID,
		ReceiverID: input.ReceiverID,
		Content:    input.Content,
	}
	if err := db.Create(&message).Error; err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отправки"})
		return
	}
	jsonMsg, err := json.Marshal(message)
	if err != nil {
		log.Printf("Ошибка сериализации сообщения: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сериализации сообщения"})
		return
	}
	sendToUser(input.ReceiverID, jsonMsg)
	sendToUser(userID, jsonMsg)
	c.JSON(http.StatusOK, message)
}

func markRead(c *gin.Context) {
	userID := c.GetInt("userID")
	var input struct {
		ReceiverID int `json:"receiver_id"`
	}
	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if err := db.Model(&Message{}).Where("sender_id = ? AND receiver_id = ? AND read_at IS NULL", input.ReceiverID, userID).
		Update("read_at", time.Now()).Error; err != nil {
		log.Printf("Ошибка отметки прочитанных: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отметки прочитанных"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Сообщения отмечены прочитанными"})
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка апгрейда WebSocket: %v", err)
		return
	}
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("Токен не предоставлен"))
		conn.Close()
		return
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		conn.WriteMessage(websocket.TextMessage, []byte("Неверный токен"))
		conn.Close()
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		conn.WriteMessage(websocket.TextMessage, []byte("Неверные данные токена"))
		conn.Close()
		return
	}
	userID := int(claims["id"].(float64))
	if oldConn, exists := clients[userID]; exists {
		oldConn.Close()
		delete(clients, userID)
	}
	clients[userID] = conn
	defer func() {
		conn.Close()
		delete(clients, userID)
	}()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Ошибка чтения WebSocket для пользователя %d: %v", userID, err)
			break
		}
	}
}

func sendToUser(userID int, message []byte) {
	if conn, ok := clients[userID]; ok {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("Ошибка отправки WebSocket сообщения пользователю %d: %v", userID, err)
			conn.Close()
			delete(clients, userID)
		}
	}
}
