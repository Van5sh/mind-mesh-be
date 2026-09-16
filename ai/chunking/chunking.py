class Chunking:
    def __init__(self, chunk_size: int):
        self.chunk_size = chunk_size

    def chunk_text(self, text: str) -> list:
        """
        Splits the input text into chunks of specified size.

        Args:
            text (str): The input text to be chunked.

        Returns:
            list: A list of text chunks.
        """
        return [text[i:i + self.chunk_size] for i in range(0, len(text), self.chunk_size)]