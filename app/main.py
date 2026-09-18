from pathlib import Path

from fastapi import FastAPI
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles

from .models import Overview
from .services import get_demo_overview

BASE_DIR = Path(__file__).resolve().parent

app = FastAPI(title="Azure Cost Cockpit", version="0.1.0")
app.mount("/static", StaticFiles(directory=BASE_DIR / "static"), name="static")


@app.get("/", response_class=FileResponse)
def dashboard() -> FileResponse:
    return FileResponse(BASE_DIR / "static" / "index.html")


@app.get("/api/overview", response_model=Overview)
def overview() -> Overview:
    return get_demo_overview()
