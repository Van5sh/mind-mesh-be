class TextEmbedding(EmbeddingBase):
    def __init__(self, model):
        super().__init__(model)

    def embed(self, text: str) -> list:
        # Implement the embedding logic specific to text using the model
        embedding = self.model.generate_text_embedding(text)
        return embedding