package slots

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterRoboticsSlots(eng *engine.Engine) {

	// ==========================================
	// SLOT: ROBOT.SENSE
	// Simulate reading from a sensor
	// ==========================================
	eng.Register("robot.sense", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		sensorType := coerce.ToString(node.Value)
		target := "sensor_data"

		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		var value interface{}
		switch sensorType {
		case "distance", "ultrasonic":
			// Simulate distance in cm (1-200)
			value = rand.Intn(200) + 1
		case "battery":
			value = rand.Intn(100)
		case "battery_voltage":
			value = 11.1 + (rand.Float64() * 1.5)
		default:
			value = "active"
		}

		slog.Info("🤖 [ROBOT] Sensing", "type", sensorType, "value", value)
		scope.Set(target, value)
		return nil
	}, engine.SlotMeta{
		Description: "Simulasi membaca data dari sensor robot.",
		Example:     "robot.sense: 'distance' { as: $d }",
	})

	// ==========================================
	// SLOT: ROBOT.ACT
	// Simulate an actuator action
	// ==========================================
	eng.Register("robot.act", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		action := coerce.ToString(node.Value)
		power := 100

		for _, c := range node.Children {
			if c.Name == "power" {
				power, _ = coerce.ToInt(parseNodeValue(c, scope))
			}
		}

		slog.Info("⚙️ [ROBOT] Actuating", "action", action, "power", fmt.Sprintf("%d%%", power))
		return nil
	}, engine.SlotMeta{
		Description: "Simulasi perintah ke aktuator robot.",
		Example:     "robot.act: 'forward' { power: 80 }",
	})
}
