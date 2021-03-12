package database

func Find{{ .ModelName }}({{ FindArgs .Columns }}) (*model.{{ .ModelName }}, error){
    db := NewDbConnection()

    result := &model.{{ .ModelName }}{}
    if err := db.First(result).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func List{{ .ModelName }}() ([]*model.{{ .ModelName }}, error){
    db := NewDbConnection()

    result := []*model.{{ .ModelName }}{}
    if err := db.First(result).Error; err != nil {
        return nil, err
    }
	return result, nil
}
