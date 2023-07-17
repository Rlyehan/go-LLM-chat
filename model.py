from fastapi import FastAPI, HTTPException, logger
from pydantic import BaseModel
from transformers import AutoTokenizer, AutoModelForCausalLM

app = FastAPI()
tokenizer = AutoTokenizer.from_pretrained("distilgpt2")
tokenizer.pad_token = tokenizer.eos_token
tokenizer.padding_side = 'left'

model = AutoModelForCausalLM.from_pretrained("distilgpt2")

class Item(BaseModel):
    text: str

@app.post("/api/query/")
async def query_model(item: Item):
    try:
        max_length = 256
        inputs = tokenizer.encode_plus(item.text, return_tensors='pt', max_length=max_length, truncation=True)
        response = model.generate(inputs['input_ids'], attention_mask=inputs['attention_mask'], max_length=max_length, temperature=0.7)
        generated = response[0][inputs["input_ids"].shape[-1]:]

        response_text = tokenizer.decode(generated, skip_special_tokens=True)
        return {"response": response_text}
    except Exception as e:
        logger.error(f"Error processing request: {e}")
        raise HTTPException(status_code=500, detail="Something went wrong.")

