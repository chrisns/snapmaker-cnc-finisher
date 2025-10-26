; GCode file with missing/malformed Snapmaker header
; This file is missing the standard Snapmaker header metadata
; Expected behavior: Issue warning but proceed with processing

G21 ; Set units to millimeters
G90 ; Absolute positioning
G94 ; Feed rate per minute

; Start cutting operations
G0 Z5.0 F1500
G0 X0 Y0
G1 Z-0.5 F600
G1 X10.0 Y0
G1 X10.0 Y10.0
G1 X0 Y10.0
G1 X0 Y0
G0 Z5.0

M5 ; Stop spindle
M2 ; End program
