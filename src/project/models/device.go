package models

type (
	Device struct {
		BaseModelWithSoftDelete
		DeviceToken string
		UserID      uint
		Token       string
	}
)

func GetDeviceByDeviceToken(deviceToken string) *Device {
	var device Device
	GetDB().Where("device_token = ?", deviceToken).First(&device)
	if device.ID == 0 {
		return nil
	}
	return &device
}

func GetDeviceByToken(token string) *Device {
	var device Device
	GetDB().Where("token = ?", token).First(&device)
	if device.ID == 0 {
		return nil
	}
	return &device
}

func (d *Device) Save() (ok bool) {
	return GetDB().Save(d).Error == nil
}

func (d *Device) GetUser() *User {
	if d.UserID == 0 {
		user := NewUser()
		d.UserID = user.ID
		return user
	}
	return GetUser(d.UserID)
}
