Set WshShell = CreateObject("WScript.Shell")
' Run the batch file completely hidden (0 = hide window)
WshShell.Run "cmd /c ""C:\Users\BASEMENT_ADMIN\NeuronFS\start_v4_swarm.bat""", 0, False
