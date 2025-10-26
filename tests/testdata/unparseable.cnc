; GCode file with unparseable/malformed lines
; Expected behavior: Skip malformed lines with warning, continue processing valid lines

;Header:
;  G54 X0 Y0 Z0
;  Tool: Snapmaker 1.5mm Flat End Mill
;MIN_Z: -1.0
;MAX_Z: 5.0

G21
G90
G94

; Valid line
G0 Z5.0 F1500
G0 X0 Y0

; Malformed line - missing required parameter format
G1 XABC Y10.0 Z-0.5 F600

; Valid line
G1 X10.0 Y10.0 Z-0.5

; Malformed line - invalid command structure
INVALID GCODE LINE HERE!!!

; Valid line
G1 X0 Y0 Z-0.5

; Another malformed line
G1 @#$% CORRUPT DATA

; Valid ending
G0 Z5.0
M5
M2
