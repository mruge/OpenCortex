# Example Dockerfile for a worker image with embedded capabilities
FROM python:3.9-slim

# Install required packages
RUN pip install pandas numpy scikit-learn

# Add capability metadata via labels (Method 1)
LABEL exec-agent.capabilities='[{"name":"python_data_processing","description":"Process data using pandas and numpy","retry_safe":true,"estimated_duration":"1m-10m"}]'
LABEL version="1.0.0"
LABEL description="Python data processing worker"
LABEL author="Smart Data Abstractor Team"
LABEL framework="python"

# Create app directory
WORKDIR /app

# Copy application code
COPY worker_script.py /app/
COPY run.sh /app/
RUN chmod +x /app/run.sh

# Add capability file (Method 2 - alternative to labels)
COPY capabilities.json /app/capabilities.json

# Set default command
CMD ["/app/run.sh"]

# Example build command:
# docker build -f worker-image.Dockerfile -t data-processor:python-v1.0 .

# The exec agent will discover this image and extract capabilities via:
# 1. Docker labels (exec-agent.capabilities)
# 2. Embedded capabilities.json file
# 3. Inference from image name and labels (framework=python)