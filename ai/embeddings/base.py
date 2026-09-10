class EmbeddingBase:
    def __init__(self, model):
        self.model = model

    def embed(self, text: str) -> list:
        # Implement the embedding logic using the model
        # For example, you can use the model to generate embeddings for the text
        embedding = self.model.generate_embedding(text)
        return embedding