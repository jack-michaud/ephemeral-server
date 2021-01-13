FROM golang:1.15.6

RUN apt update && apt install -y unzip golang rsync python-pip
RUN pip install ansible

# Install terraform
RUN wget https://releases.hashicorp.com/terraform/0.14.4/terraform_0.14.4_linux_arm.zip -O /tmp/terraform.zip
RUN unzip /tmp/terraform.zip
RUN rm /tmp/terraform.zip
RUN mv terraform /bin/terraform

WORKDIR /code

ADD . /code/

RUN make
CMD ["/code/build/ephemeralbot"]
