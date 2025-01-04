func validateToken(tokenString string) (*CustomClaims, error) {
    secret := []byte("your-secret-key")
    claims := &CustomClaims{}

    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return secret, nil
    })
    if err != nil || !token.Valid {
        return nil, err
    }

    // Additional checks
    if claims.ExpiresAt < time.Now().Unix() {
        return nil, fmt.Errorf("token has expired")
    }
    
    // Check user roles/permissions if necessary
    // if claims.Role != "admin" {
    //     return nil, fmt.Errorf("user is not authorized")
    // }

    return claims, nil
}