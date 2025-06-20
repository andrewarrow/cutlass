import torchaudio as ta
from chatterbox.tts import ChatterboxTTS
import sys
import glob
import random 

model = ChatterboxTTS.from_pretrained(device="mps")

text = sys.argv[1]
file = sys.argv[2]
voice = sys.argv[3]
print(text)
print(file)

AUDIO_PROMPT_PATH = f"/Users/aa/cs/voices/{voice}.wav"
print(f"Using voice: {AUDIO_PROMPT_PATH}")
wav = model.generate(text, 
    audio_prompt_path=AUDIO_PROMPT_PATH,
    #exaggeration=1.2,
    #cfg_weight=0.3, 
    #temperature=0.9  
    )
ta.save(file, wav, model.sr)
