func validateToken(tokenString string) (string, error) {
    secret := []byte("your-secret-key")
    token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.HS256); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return secret, nil
    })

    if err != nil || !token.Valid {
        return "", err
    }

    claims, ok := token.Claims.(*CustomClaims)
    if !ok {
        return "", err
    }

    if time.Now().After(claims.Exp) {
        return "", errors.New("token expired")
    }

    if claims.Issuer != "your-issuer" {
        return "", errors.New("invalid issuer")
    }

    if claims.Audience != "your-audience" {
        return "", errors.New("invalid audience")
    }

    return claims.UserID, nil
}