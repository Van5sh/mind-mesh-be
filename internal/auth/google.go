package auth

func (s *Service) GoogleLogin(code string) (string, error) {
	// Exchange the authorization code for an access token
	token, err := s.googleOAuthConfig.Exchange(s.ctx, code)
	if err != nil {
		return "", err
	}

	// Use the access token to get user info from Google
}
