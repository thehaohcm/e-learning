package study

import "testing"

func TestDeviceSettings(t *testing.T) {
	d := ParseDeviceSettings([]byte(`{"subject_ids":[2,2,-1],"device_mode":"bad","reminder_every_min":3,"active_from":"7:00","active_to":"21:30"}`))
	if len(d.SubjectIDs) != 1 || d.SubjectIDs[0] != 2 {
		t.Fatalf("subject_ids = %v", d.SubjectIDs)
	}
	if d.DeviceMode != "" || d.ReminderEveryMin != 0 || d.ActiveFrom != "" || d.ActiveTo != "21:30" {
		t.Fatalf("not sanitized: %+v", d)
	}
	if p := ParseDeviceSettings([]byte(`{"popup_cards":7}`)); p.PopupCards != 0 {
		t.Fatalf("popup_cards 7 should be dropped: %+v", p)
	}
	if a := ParseSettings([]byte(`{"popup_cards":3}`)); a.PopupCards != 10 {
		t.Fatalf("account popup_cards default: %d", a.PopupCards)
	}
	s := DefaultSettings().ForDevice(DeviceSettings{DeviceMode: DeviceModeQuiz, ReminderEveryMin: 20, PopupCards: 20})
	if s.DeviceMode != DeviceModeQuiz || s.ReminderEveryMin != 20 || s.ActiveFrom != "07:30" || s.PopupCards != 20 {
		t.Fatalf("ForDevice: %+v", s)
	}
	if e := ParseDeviceSettings(nil); e.SubjectIDs == nil || len(e.SubjectIDs) != 0 {
		t.Fatalf("empty: %+v", e)
	}
}

func TestExtraNewLeft(t *testing.T) {
	s := DefaultSettings()
	if s.ExtraNewLeft("2026-10-23") != MaxExtraNew {
		t.Fatal("fresh day")
	}
	s.ExtraNewDate, s.ExtraNew = "2026-10-23", 15
	if s.ExtraNewLeft("2026-10-23") != 5 || s.ExtraNewLeft("2026-10-24") != MaxExtraNew {
		t.Fatal("left")
	}
	s.ExtraNew = 25
	if s.ExtraNewLeft("2026-10-23") != 0 {
		t.Fatal("never negative")
	}
}
