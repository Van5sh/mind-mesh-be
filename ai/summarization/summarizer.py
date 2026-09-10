class Summarizer:
    def __init__(self, model):
        self.model = model
    def summarize(self,text: str)-> str:
        # Implement the summarization logic using the model
        # For example, you can use the model to generate a summary of the text
        summary = self.model.generate_summary(text)
        return summary