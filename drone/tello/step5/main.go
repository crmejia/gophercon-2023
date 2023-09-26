package main

import (
	"fmt"
	"time"

	"sync/atomic"

	gobot "gobot.io/x/gobot/v2"
	"gobot.io/x/gobot/v2/platforms/dji/tello"
	"gobot.io/x/gobot/v2/platforms/joystick"
)

var drone = tello.NewDriver("8888")

type pair struct {
	x float64
	y float64
}

var leftX, leftY, rightX, rightY atomic.Value

const offset = 32767.0

func main() {
	var joystickAdaptor = joystick.NewAdaptor("0")
	var stick = joystick.NewDriver(joystickAdaptor, joystick.Dualshock4)
	var currentFlightData *tello.FlightData

	work := func() {
		leftX.Store(float64(0.0))
		leftY.Store(float64(0.0))
		rightX.Store(float64(0.0))
		rightY.Store(float64(0.0))

		configureStickEvents(stick)
		fmt.Println("takeoff...")

		drone.On(tello.FlightDataEvent, func(data interface{}) {
			fd := data.(*tello.FlightData)
			currentFlightData = fd
		})

		gobot.After(20*time.Second, func() {
			drone.Land()
		})

		gobot.Every(1*time.Second, func() {
			printFlightData(currentFlightData)
		})

		gobot.Every(50*time.Millisecond, func() {
			rightStick := getRightStick()

			switch {
			case rightStick.y < -10:
				drone.Forward(tello.ValidatePitch(rightStick.y, offset))
			case rightStick.y > 10:
				drone.Backward(tello.ValidatePitch(rightStick.y, offset))
			default:
				drone.Forward(0)
			}

			switch {
			case rightStick.x > 10:
				drone.Right(tello.ValidatePitch(rightStick.x, offset))
			case rightStick.x < -10:
				drone.Left(tello.ValidatePitch(rightStick.x, offset))
			default:
				drone.Right(0)
			}
		})

		gobot.Every(50*time.Millisecond, func() {
			leftStick := getLeftStick()
			switch {
			case leftStick.y < -10:
				drone.Up(tello.ValidatePitch(leftStick.y, offset))
			case leftStick.y > 10:
				drone.Down(tello.ValidatePitch(leftStick.y, offset))
			default:
				drone.Up(0)
			}

			switch {
			case leftStick.x > 20:
				drone.Clockwise(tello.ValidatePitch(leftStick.x, offset))
			case leftStick.x < -20:
				drone.CounterClockwise(tello.ValidatePitch(leftStick.x, offset))
			default:
				drone.Clockwise(0)
			}
		})
	}

	robot := gobot.NewRobot("tello",
		[]gobot.Connection{joystickAdaptor},
		[]gobot.Device{drone, stick},
		work,
	)

	robot.Start()
}

func configureStickEvents(stick *joystick.Driver) {
	stick.On(joystick.TrianglePress, func(data interface{}) {
		drone.TakeOff()
	})

	stick.On(joystick.XPress, func(data interface{}) {
		drone.Land()
	})

	stick.On(joystick.UpPress, func(data interface{}) {
		fmt.Println("FrontFlip")
		drone.FrontFlip()
	})

	stick.On(joystick.DownPress, func(data interface{}) {
		fmt.Println("BackFlip")
		drone.BackFlip()
	})

	stick.On(joystick.RightPress, func(data interface{}) {
		fmt.Println("RightFlip")
		drone.RightFlip()
	})

	stick.On(joystick.LeftPress, func(data interface{}) {
		fmt.Println("LeftFlip")
		drone.LeftFlip()
	})

	stick.On(joystick.LeftX, func(data interface{}) {
		val := float64(data.(int))
		leftX.Store(val)
	})

	stick.On(joystick.LeftY, func(data interface{}) {
		val := float64(data.(int))
		leftY.Store(val)
	})

	stick.On(joystick.RightX, func(data interface{}) {
		val := float64(data.(int))
		rightX.Store(val)
	})

	stick.On(joystick.RightY, func(data interface{}) {
		val := float64(data.(int))
		rightY.Store(val)
	})
}

func printFlightData(d *tello.FlightData) {
	if d.BatteryLow {
		fmt.Printf(" -- Battery low: %d%% --\n", d.BatteryPercentage)
	}

	displayData := `
Battery:		%d%%
Height:         %d
Ground Speed:   %d

`
	fmt.Printf(displayData, d.BatteryPercentage, d.Height, d.GroundSpeed)
}

func getLeftStick() pair {
	s := pair{x: 0, y: 0}
	s.x = leftX.Load().(float64)
	s.y = leftY.Load().(float64)
	return s
}

func getRightStick() pair {
	s := pair{x: 0, y: 0}
	s.x = rightX.Load().(float64)
	s.y = rightY.Load().(float64)
	return s
}
