package lightify

func Init() error {
	LightifyConfig = NewConfig()

	token, err := GenerateToken()
	if err != nil {
		return err
	}

	lightifyToken = token

	go RefreshTokenRoutine()

	return nil
}
