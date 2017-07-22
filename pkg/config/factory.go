package config

func Create() (Interface, error) {
	config := new(configuration)
	if err := config.init(); err != nil{
		return nil, err
	}
	return config, nil
}
