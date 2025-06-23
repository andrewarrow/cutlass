go build
rm trip.fcpxml

./cutlass fcp add-txt "Hey u there?" -o trip.fcpxml
./cutlass fcp add-txt -i trip.fcpxml "Yes, I'm here." --original-text "hey u there?" -o trip.fcpxml
./cutlass fcp add-txt -i trip.fcpxml "u sure?" -o trip.fcpxml
#./cutlass fcp add-txt -i trip.fcpxml "ok no problem!" --original-text "teach me vibe coding" -o trip.fcpxml




#./cutlass fcp add-video ~/Documents/test1.mov -o trip.fcpxml
#./cutlass fcp add-pip-video ~/Desktop/test1.mov --input trip.fcpxml -o trip.fcpxml
#./cutlass fcp add-image ./assets/waymo.png --duration 28 -o filename.fcpxml
#./cutlass fcp add-slide 0 --input filename.fcpxml -o filename.fcpxml
#./cutlass fcp add-text slide_text.txt --input filename.fcpxml -o filename.fcpxml
#./cutlass fcp add-image ./assets/waymo.png --duration 28 --input filename.fcpxml -o filename.fcpxml
#./cutlass fcp add-slide 28 --input filename.fcpxml -o filename.fcpxml
#./cutlass fcp add-text slide_text2.txt --offset 28 --input filename.fcpxml -o filename.fcpxml
#./cutlass fcp add-audio ./data/waymo_audio/output.wav --input filename.fcpxml -o filename.fcpxml


