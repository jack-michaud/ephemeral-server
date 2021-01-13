FROM golang:1.15.6

ARG TARGETPLATFORM

RUN adduser --home /home/eph --gecos "" --disabled-password eph


RUN apt update && apt install -y unzip golang rsync python-pip
RUN pip install ansible

# Install terraform
RUN echo "$TARGETPLATFORM"
RUN if [ "$TARGETPLATFORM" = "linux/386" ] ; then \
    export TERRAFORM_URL="https://releases.hashicorp.com/terraform/0.14.4/terraform_0.14.4_linux_386.zip"; \
  elif [ "$TARGETPLATFORM" = "linux/amd64" ] ; then \
    export TERRAFORM_URL="https://releases.hashicorp.com/terraform/0.14.4/terraform_0.14.4_linux_amd64.zip"; \
  elif [ "$TARGETPLATFORM" = "linux/arm/v7" ] ; then \
    export TERRAFORM_URL="https://releases.hashicorp.com/terraform/0.14.4/terraform_0.14.4_linux_arm.zip"; \
  elif [ "$TARGETPLATFORM" = "linux/arm64" ] ; then \
    export TERRAFORM_URL="https://releases.hashicorp.com/terraform/0.14.4/terraform_0.14.4_linux_arm64.zip"; \
  else \
  exit 1; \
  fi; \
  wget $TERRAFORM_URL -O /tmp/terraform.zip
RUN unzip /tmp/terraform.zip
RUN rm /tmp/terraform.zip
RUN mv terraform /bin/terraform

WORKDIR /code

RUN chown -R eph:eph /code

USER eph

ADD . /code/

RUN make
CMD ["/code/build/ephemeralbot"]

