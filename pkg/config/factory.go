package config


func Create() (Interface, error) {
	config := new(configuration)
	config.init()
	return config, nil
}
