from pathlib import Path
import logging

from azure.core.exceptions import AzureError
from fastapi import FastAPI, HTTPException
from fastapi.responses import FileResponse
from fastapi.staticfiles import StaticFiles

from .models import Overview
from .services import get_overview

BASE_DIR = Path(__file__).resolve().parent

app = FastAPI(title="Azure Cost Cockpit", version="0.1.0")
app.mount("/static", StaticFiles(directory=BASE_DIR / "static"), name="static")
logger = logging.getLogger(__name__)


@app.get("/", response_class=FileResponse)
def dashboard() -> FileResponse:
    return FileResponse(BASE_DIR / "static" / "index.html")


@app.get("/api/overview", response_model=Overview)
def overview() -> Overview:
    try:
        return get_overview()
    except ValueError as error:
        raise HTTPException(status_code=503, detail=str(error)) from error
    except AzureError as error:
        logger.exception("Azure Cost Management request failed")
        if getattr(error, "status_code", None) == 429:
            detail = "Azure Cost Management is rate limiting requests. Please wait and retry."
        else:
            detail = "Azure Cost Management is temporarily unavailable. Please retry shortly."
        raise HTTPException(
            status_code=503,
            detail=detail,
        ) from error
