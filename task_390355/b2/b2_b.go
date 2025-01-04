type CustomClaims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Exp    time.Time `json:"exp"`
    jwt.StandardClaims
}