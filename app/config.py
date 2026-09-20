import json
from dataclasses import dataclass
from pathlib import Path


@dataclass(frozen=True)
class Settings:
    data_source: str
    subscription_id: str | None


def get_settings() -> Settings:
    config_path = Path(__file__).resolve().parent.parent / "cockpit.config.json"
    try:
        config = json.loads(config_path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        config = {}
    data_source = str(config.get("data_source", "demo")).lower()
    subscription_id = config.get("azure_subscription_id") or None
    if data_source not in {"demo", "azure"}:
        raise ValueError("COCKPIT_DATA_SOURCE must be either 'demo' or 'azure'")
    return Settings(data_source=data_source, subscription_id=subscription_id)
