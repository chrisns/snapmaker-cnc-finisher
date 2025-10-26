; Snapmaker Finishing Pass Test File - 4-Axis
; Header Start
;MIN_Z: -3.0
;MAX_Z: 0.5
;TOOL: 2mm Ball End Mill
;ROTARY: B-axis enabled
;Header End

; Initialize spindle
M3 S1200

; Rapid to start position with B-axis
G0 X0 Y0 Z5.0 B0

; Shallow pass with rotation (should be removed with 1.0mm allowance)
G1 Z-0.3 F1000
G1 X10.0 Y0 B15.0 F1500
G1 X10.0 Y5.0 B30.0
G1 X0 Y5.0 B45.0
G1 X0 Y0 B60.0

; Rapid move with B-axis positioning
G0 Z5.0
G0 X20.0 Y0 B0

; Deep pass with rotation (should be kept)
G1 Z-2.0 F1000
G1 X30.0 Y0 B20.0 F1500
G1 X30.0 Y5.0 B40.0
G1 X20.0 Y5.0 B60.0
G1 X20.0 Y0 B80.0

; Another shallow multi-axis move
G0 Z5.0
G0 X40.0 Y0 B90.0
G1 Z-0.4 F1000
G1 X50.0 Y0 B105.0 F1500
G1 X50.0 Y5.0 B120.0

; Deep cut with complex rotation
G0 Z5.0
G0 X0 Y10.0 B0
G1 Z-2.5 F1000
G1 X10.0 Y10.0 B30.0 F1500
G1 X10.0 Y15.0 B60.0
G1 X0 Y15.0 B90.0

; Shallow pass - edge case
G0 Z5.0
G0 X20.0 Y10.0 B120.0
G1 Z-0.9 F1000
G1 X30.0 Y10.0 B135.0 F1500
G1 X30.0 Y15.0 B150.0

; Very deep finishing
G0 Z5.0
G0 X40.0 Y10.0 B180.0
G1 Z-2.8 F1000
G1 X50.0 Y10.0 B210.0 F1500
G1 X50.0 Y15.0 B240.0

; Rapid moves preserving B position
G0 Z10.0
G0 X0 Y0 B0

; Stop spindle
M5

; End of program
M2
