; Snapmaker Finishing Pass Test File - 3-Axis
; Header Start
;MIN_Z: -5.0
;MAX_Z: 0.0
;TOOL: 3mm End Mill
;Header End

; Initialize spindle
M3 S1000

; Rapid to start position
G0 X0 Y0 Z5.0

; Shallow pass (should be removed with 1.0mm allowance)
G1 Z-0.2 F1000
G1 X10.0 Y0 F1500
G1 X10.0 Y10.0
G1 X0 Y10.0
G1 X0 Y0

; Rapid move to next position
G0 Z5.0
G0 X20.0 Y0

; Deep pass (should be kept)
G1 Z-2.5 F1000
G1 X30.0 Y0 F1500
G1 X30.0 Y10.0
G1 X20.0 Y10.0
G1 X20.0 Y0

; Rapid move
G0 Z5.0
G0 X40.0 Y0

; Another shallow pass
G1 Z-0.5 F1000
G1 X50.0 Y0 F1500
G1 X50.0 Y10.0
G1 X40.0 Y10.0
G1 X40.0 Y0

; Rapid move
G0 Z5.0
G0 X60.0 Y0

; Medium depth pass (edge case around 1.0mm)
G1 Z-1.2 F1000
G1 X70.0 Y0 F1500
G1 X70.0 Y10.0
G1 X60.0 Y10.0
G1 X60.0 Y0

; Shallow pass again
G0 Z5.0
G0 X80.0 Y0
G1 Z-0.3 F1000
G1 X90.0 Y0 F1500
G1 X90.0 Y10.0
G1 X80.0 Y10.0
G1 X80.0 Y0

; Deep finishing pass
G0 Z5.0
G0 X0 Y20.0
G1 Z-3.0 F1000
G1 X10.0 Y20.0 F1500
G1 X10.0 Y30.0
G1 X0 Y30.0
G1 X0 Y20.0

; Very shallow (< 0.5mm)
G0 Z5.0
G0 X20.0 Y20.0
G1 Z-0.1 F1000
G1 X30.0 Y20.0 F1500
G1 X30.0 Y30.0
G1 X20.0 Y30.0
G1 X20.0 Y20.0

; Another deep pass
G0 Z5.0
G0 X40.0 Y20.0
G1 Z-4.0 F1000
G1 X50.0 Y20.0 F1500
G1 X50.0 Y30.0
G1 X40.0 Y30.0
G1 X40.0 Y20.0

; Mixed depth - shallow Z moves
G0 Z5.0
G0 X60.0 Y20.0
G1 Z-0.8 F1000
G1 X70.0 Y20.0 F1500
G1 X70.0 Y30.0
G1 X60.0 Y30.0
G1 X60.0 Y20.0

; Return to safe height and stop spindle
G0 Z10.0
M5

; End of program
M2
