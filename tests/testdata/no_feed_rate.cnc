; GCode file with G1 commands missing F parameter entirely
; Expected behavior: Use default feed rate (1000 mm/min) with warning

;Header:
;  G54 X0 Y0 Z0
;  Tool: Snapmaker 1.5mm Flat End Mill
;MIN_Z: -1.0
;MAX_Z: 5.0

G21
G90
G94

; Rapid move to start position (no feed rate needed)
G0 Z5.0
G0 X0 Y0

; Cutting moves with NO feed rate specified anywhere
G1 Z-0.5
G1 X10.0 Y0
G1 X10.0 Y10.0 Z-0.5
G1 X0 Y10.0
G1 X0 Y0

; Rapid move back up
G0 Z5.0

; Another cutting pass - still no feed rate
G1 X20.0 Y0 Z-0.5
G1 X20.0 Y10.0

; Ending
G0 Z5.0
M5
M2
