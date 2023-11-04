package baby

// Baby - baby info (matching the Nanit API)
type Baby struct {
	UID       string `json:"uid"`
	Name      string `json:"name"`
	CameraUID string `json:"camera_uid"`
}
