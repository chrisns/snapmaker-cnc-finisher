; All Deep Test File - Minimum Filtering Test
; Header Start
;MIN_Z: -5.0
;MAX_Z: 0.0
;TOOL: 6mm End Mill
;Header End

M3 S1000

G0 X0 Y0 Z5.0

; All these should be kept with 1.0mm allowance
G1 Z-1.5 F800
G1 X5.0 Y0 F1000

G1 Z-2.0 F800
G1 X10.0 Y0 F1000

G1 Z-2.5 F800
G1 X15.0 Y0 F1000

G1 Z-3.0 F800
G1 X20.0 Y0 F1000

G1 Z-3.5 F800
G1 X25.0 Y0 F1000

G1 Z-4.0 F800
G1 X30.0 Y0 F1000

G1 Z-4.5 F800
G1 X35.0 Y0 F1000

G1 Z-5.0 F800
G1 X40.0 Y0 F1000

; Also test exact boundary (just over 1.0mm)
G1 Z-1.1 F800
G1 X45.0 Y0 F1000

G1 Z-1.01 F800
G1 X50.0 Y0 F1000

G0 Z10.0
M5
M2
