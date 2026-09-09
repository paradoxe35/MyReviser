package ui

import (
	"fyne.io/fyne/v2"
	"github.com/paradoxe35/encre/internal/logger"
)

// NotificationManager handles user notifications
type NotificationManager struct {
	app fyne.App
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(app fyne.App) *NotificationManager {
	return &NotificationManager{
		app: app,
	}
}

// ShowSuccess shows a success notification
func (n *NotificationManager) ShowSuccess(title, content string) {
	notification := fyne.NewNotification(title, content)
	n.app.SendNotification(notification)
	logger.Info("Success notification shown", "title", title, "content", content)
}

// ShowError shows an error notification
func (n *NotificationManager) ShowError(title, content string) {
	notification := fyne.NewNotification(title, content)
	n.app.SendNotification(notification)
	logger.Error("Error notification shown", "title", title, "content", content)
}

// ShowInfo shows an info notification
func (n *NotificationManager) ShowInfo(title, content string) {
	notification := fyne.NewNotification(title, content)
	n.app.SendNotification(notification)
	logger.Info("Info notification shown", "title", title, "content", content)
}
