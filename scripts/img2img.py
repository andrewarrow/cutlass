from diffusers import StableDiffusion3Img2ImgPipeline
import torch
from PIL import Image

model_id = "stabilityai/stable-diffusion-3.5-medium"

pipeline = StableDiffusion3Img2ImgPipeline.from_pretrained(
    model_id,
    torch_dtype=torch.bfloat16
)

# Use MPS on Mac
import platform
pipeline = pipeline.to("mps")
#pipeline.enable_model_cpu_offload()
pipeline.enable_attention_slicing()

# Load the input image
input_image = Image.open("k.jpg")

prompt = "character standing upright, full body visible, complete legs and feet drawn, no desk, no furniture, clean background, anatomically correct proportions"
negative_prompt = "sitting, desk, furniture, chair, deformed legs, missing legs, cropped body, partial body"

image = pipeline(
    prompt=prompt,  # Text description of desired output
    image=input_image,  # Input image to transform
    negative_prompt=negative_prompt,  # What to avoid in the output
    num_inference_steps=30,  # Number of denoising steps (1-1000, higher = better quality but slower)
    guidance_scale=12.0,  # How closely to follow prompt (1.0-20.0, higher = more adherence to prompt)
    strength=0.8,  # How much to change input image (0.0-1.0, higher = more transformation)
    max_sequence_length=256,  # Maximum token length for prompt processing (77-512)
    # Optional parameters you can add:
    # generator=torch.Generator().manual_seed(42),  # For reproducible results
    # eta=0.0,  # Controls noise schedule (0.0-1.0)
    # guidance_rescale=0.0,  # Rescale guidance to prevent over-exposure (0.0-1.0)
    # clip_skip=None,  # Skip layers in CLIP text encoder (None or 1-12)
).images[0]
image.save("whimsical.png")
