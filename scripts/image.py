from diffusers import StableDiffusion3Pipeline
import torch

model_id = "stabilityai/stable-diffusion-3.5-medium"

pipeline = StableDiffusion3Pipeline.from_pretrained(
    model_id,
    torch_dtype=torch.bfloat16
)

# Use MPS on Mac
import platform
pipeline = pipeline.to("mps")
#pipeline.enable_model_cpu_offload()
pipeline.enable_attention_slicing()

prompt = "a logo of a cutlass sword slicing thru real world xml tags that are real physical objects"

image = pipeline(
    prompt=prompt,
    num_inference_steps=20,
    guidance_scale=3.0,
    max_sequence_length=256,
    height=512,
    width=512,
).images[0]
image.save("whimsical.png")
