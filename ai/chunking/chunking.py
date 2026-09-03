from dataclasses import dataclass
import re


@dataclass
class Chunk:
    text: str
    index: int
    start_char: int
    end_char: int


def _normalize(text: str) -> str:
    text = text.replace("\r\n", "\n").replace("\r", "\n")
    text = re.sub(r"[ \t]+", " ", text)
    text = re.sub(r"\n{3,}", "\n\n", text)
    return text.strip()


def chunk_text(
    text: str,
    chunk_size: int = 900,
    overlap: int = 150,
) -> list[Chunk]:
    if chunk_size <= 0:
        raise ValueError("chunk_size must be > 0")
    if overlap < 0 or overlap >= chunk_size:
        raise ValueError("overlap must be >= 0 and < chunk_size")

    text = _normalize(text)
    if not text:
        return []

    chunks = []
    start = 0
    index = 0

    while start < len(text):
        target_end = min(start + chunk_size, len(text))
        end = target_end

        if target_end < len(text):
            # Prefer a natural boundary near the target.
            for separator in ("\n\n", "\n", ". ", " ", ""):
                candidate = text.rfind(separator, start + chunk_size // 2, target_end)
                if candidate > start:
                    end = candidate + (len(separator) if separator else 0)
                    break

        piece = text[start:end].strip()
        if piece:
            actual_start = text.find(piece, start, end)
            actual_end = actual_start + len(piece)
            chunks.append(
                Chunk(
                    text=piece,
                    index=index,
                    start_char=actual_start,
                    end_char=actual_end,
                )
            )
            index += 1

        if end >= len(text):
            break

        start = max(end - overlap, start + 1)

    return chunks