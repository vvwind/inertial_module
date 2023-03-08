package main

import (
	"fmt"
	"machine"
	"math"
	"time"
)

const (
	offx = 31
	offy = 47
	offz = -23
)

func acc(dataxyz []byte) []float64 {
	x := (int(dataxyz[0]) | int(dataxyz[1])<<8>>2) + offx
	y := (int(dataxyz[2]) | int(dataxyz[3])<<8>>2) + offy
	z := (int(dataxyz[4]) | int(dataxyz[5])<<8>>2) + offz
	x1 := float64(x) / 4096.0
	y1 := float64(y) / 4096.0
	z1 := float64(z) / 4096.0
	return []float64{x1, y1, z1}
}

func findAngle(ax float64, ay float64, az float64) []float64 {
	xAngle := math.Atan(ax / math.Pow(math.Pow(ay, 2)+math.Pow(az, 2), 0.5))
	yAngle := math.Atan(ay / math.Pow(math.Pow(ax, 2)+math.Pow(az, 2), 0.5))
	zAngle := math.Atan((math.Pow(math.Pow(ay, 2)+math.Pow(ax, 2), 0.5)) / az)
	xAngle *= 180
	yAngle *= 180
	zAngle *= 180
	xAngle /= math.Pi
	yAngle /= math.Pi
	zAngle /= math.Pi
	return []float64{xAngle, yAngle, zAngle}
}

func main() {

	uartGSM := machine.UART1
	uartGSM.Configure(machine.UARTConfig{
		BaudRate: 9600,
		RX:       machine.UART_RX_PIN,
		TX:       machine.UART_TX_PIN,
	})

	uart := machine.UART2
	uart.Configure(machine.UARTConfig{
		BaudRate: 115200,
		RX:       machine.A3,
		TX:       machine.A2,
	})

	i2c := machine.I2C0
	err := i2c.Configure(machine.I2CConfig{
		SCL: machine.I2C0_SCL_PIN,
		SDA: machine.I2C0_SDA_PIN,
	})
	if err != nil {
		println("could not configure I2C:", err)
		return
	}

	time.Sleep(0 * time.Second)

	//uart.Configure(machine.UARTConfig{TX: tx, RX: rx})
	////data := make([]byte, 1)
	////uart.Write([]byte(prompt))
	//if uart.Buffered() > 0 {
	//	discard, _ := uart.ReadByte()
	//	println("discarded", discard)
	//}

	r := make([]byte, 3)
	machine.I2C0.Tx(uint16(0x0D), []byte{0x00}, r)
	machine.I2C0.Tx(uint16(0x10), []byte{0xB6}, r)

	uartGSM.Write([]byte("AT"))
	if uartGSM.Buffered() > 0 {
		dummy := []byte{}
		for i := 0; i < uartGSM.Buffered(); i++ {
			mbyte, merr := uartGSM.ReadByte()
			if merr != nil {
				uart.Write([]byte("Error:\n\r"))
				uart.Write([]byte(fmt.Sprintf("%e", merr)))
				uart.Write([]byte("\n\r"))
			} else {
				dummy = append(dummy, mbyte)
			}

		}
		uart.Write([]byte("Init AT:\n\r"))
		uart.Write(dummy)
		uart.Write([]byte("\n\r"))
	}

	uartGSM.Write([]byte("AT+CGNSPWR=1"))
	if uartGSM.Buffered() > 0 {
		dummy := []byte{}
		for i := 0; i < uartGSM.Buffered(); i++ {
			mbyte, merr := uartGSM.ReadByte()
			if merr != nil {
				uart.Write([]byte("Error:\n\r"))
				uart.Write([]byte(fmt.Sprintf("%e", merr)))
				uart.Write([]byte("\n\r"))
			} else {
				dummy = append(dummy, mbyte)
			}

		}
		uart.Write([]byte("StartingUp :\n\r"))
		uart.Write(dummy)
		uart.Write([]byte("\n\r"))
	}
	time.Sleep(3 * time.Second)
	for {
		dataxyz := make([]byte, 6)
		machine.I2C0.Tx(uint16(0x38), []byte{0x02}, dataxyz)
		time.Sleep(1 * time.Second)
		my_data := make([]float64, 3)
		my_data = acc(dataxyz)
		angles := make([]float64, 3)
		angles = findAngle(my_data[0], my_data[1], my_data[2])
		sta1 := fmt.Sprintf("%f", my_data[0])
		sta2 := fmt.Sprintf("%f", my_data[1])
		sta3 := fmt.Sprintf("%f", my_data[2])
		x := []byte(sta1)
		y := []byte(sta2)
		z := []byte(sta3)
		uart.Write([]byte("=============\n\r"))
		uart.Write([]byte("x = "))
		uart.Write(x)
		uart.Write([]byte(" g "))
		uart.Write([]byte(", X_Angle =  "))
		uart.Write([]byte(fmt.Sprintf("%f", angles[0])))

		uart.Write([]byte("\n\r"))
		println("")
		uart.Write([]byte("y = "))
		uart.Write(y)
		uart.Write([]byte(" g "))
		uart.Write([]byte(", Y_Angle =  "))
		uart.Write([]byte(fmt.Sprintf("%f", angles[1])))

		uart.Write([]byte("\n\r"))
		println("")
		uart.Write([]byte("z = "))
		uart.Write(z)
		uart.Write([]byte(" g "))
		uart.Write([]byte(", Z_Angle =  "))
		uart.Write([]byte(fmt.Sprintf("%f", angles[2])))

		uart.Write([]byte("\n\r"))
		println("")
		uart.Write([]byte("=============\n\r"))

		time.Sleep(1 * time.Second)
		// Wait for the response
		uartGSM.Write([]byte("AT+CGNSINF"))
		time.Sleep(1 * time.Second)

		// Read the response
		mbytes := []byte{}
		if uartGSM.Buffered() > 0 {
			for i := 0; i < uartGSM.Buffered(); i++ {
				mbyte, merr := uartGSM.ReadByte()
				if merr != nil {
					uart.Write([]byte("Error:\n\r"))
					uart.Write([]byte(fmt.Sprintf("%e", merr)))
					uart.Write([]byte("\n\r"))
				} else {
					mbytes = append(mbytes, mbyte)
				}
			}
		}
		//my_str := string(mbytes)[:]
		//dummy2 := strings.Split(my_str, " ")
		uart.Write([]byte("Got:\n\r"))
		uart.Write(mbytes)
		uart.Write([]byte("\n\r"))

		//uartGSM.Write([]byte("AT"))
		//mbyte, merr := uartGSM.ReadByte()
		//println(mbyte)
		//if merr != nil {
		//	println("err: ", merr)
		//}
		//uartGSM.Write([]byte("AT+CPIN"))
		//mbyte, merr = uartGSM.ReadByte()
		//println(mbyte)
		//if merr != nil {
		//	println("err: ", merr)
		//}
	}

	//for {
	//	for {
	//		println(errz)
	//		uart.WriteByte(dataxyz[3])
	//		println("first: ", dataxyz[3])
	//		time.Sleep(1 * time.Second)
	//		inByte, _ := uart.ReadByte()
	//		println("scond: ", inByte)
	//		if inByte != byte(13) {
	//			continue
	//		} else {
	//			break
	//		}

	//uart.WriteByte(dataxyz[1])
	//
	//x, errwr := uart.ReadByte()
	//if errwr != nil {
	//	println(errwr)
	//}
	//println("x = ", x)
	//uart.WriteByte(dataxyz[3])
	//
	//y, _ := uart.ReadByte()
	//println("y = ", y)
	//
	//uart.WriteByte(dataxyz[5])
	//
	//z, _ := uart.ReadByte()
	//println("z = ", z)

	//}
	//for {
	//	for {
	//		if uart.Buffered() > 0 {
	//			inByte, _ := uart.ReadByte()
	//			print(string(inByte))
	//			uart.WriteByte(inByte)
	//			if inByte != byte(13) {
	//				data = append(data, inByte)
	//				continue
	//			} else {
	//				break
	//			}
	//		}
	//
	//	}
	//
	//	time.Sleep(10 * time.Millisecond)
	//	//output := "\r\nHello " + string(data) + "\r\n"
	//
	//	dataxyz := make([]byte, 6)
	//
	//	var datacc []uint16
	//	machine.I2C0.Tx(uint16(0x38), []byte{0x02}, dataxyz)
	//	for i := 0; i < 3; i++ {
	//		temp := uint16(dataxyz[i])
	//		vart := uint16(dataxyz[i+1])
	//		temp = temp << 8
	//		temp = temp | vart
	//		temp = temp >> 2
	//		datacc = append(datacc, temp)
	//	}
	//	//a_X := 180 * math.Atan(float64(data[0]) / math.Sqrt2(float64(datacc[0]*datacc[0])+float64(datacc[2]*datacc[2])) / math.Pi
	//	println("==============")
	//	println((dataxyz[1]))
	//	println((dataxyz[3]))
	//	println((dataxyz[5]))
	//	println("==============")
	//	println("==============")
	//	println((datacc[0]))
	//	println((datacc[1]))
	//	println((datacc[2]))
	//	println("==============")
	//
	//	data = nil
	//}
	//}
}
