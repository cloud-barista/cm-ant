# CM-ANT 클라우드 마이그레이션 전환 상태 검증 프레임워크
###
```text
🧨 [WARNING]
🧨 CM-ANT is currently under development.
🧨 So, we do not recommend using the current release in production.
🧨 Please note that the functionalities of CM-ANT are not stable and secure yet.
🧨 If you have any difficulties in using CM-ANT, please let us know.
🧨 (Open an issue or Join the Cloud-Migrator Slack)
```
  
<br/>

# Overview
클라우드 마이그레이션 전환 상태 검증 프레임워크는 클라우드로 전환하는 절차(이하 마이그레이션) 전후로 성능 혹은 가격 및 비용에 대한 검증 기능을 수행하는 프레임워크이다.

크게 두가지 범주의 기능을 제공하며 각각은 다음과 같다.
- 목표 클라우드 인프라 전환 비용 검증 및 예측
- 목표 클라우드 인프라 온디맨드 성능 평가 및 검증

목표 클라우드 인프라 전환 비용 검증 및 예측은 `1)` 마이그레이션이 진행되기 전 추천하는 혹은 목표하는 스펙의 인프라에 대한 가격 정보를 제공한다. 또한 `2)` 특정 csp 의 운영 비용 정보 제공 및 `3)` 예측 비용에 대한 정보를 제공한다.

목표 클라우드 인프라 온디맨드 성능 평가 및 검증은 `1)` 마이그레이션 된 인프라에서 작동하는 애플리케이션에 대해 성능 평가를 진행한다. 진행한 성능 평가 결과를 바탕으로 `2)` 사용자에게 마이그레이션 성능 검증 정보를 제공한다.

각 범주의 기능은 다른 서브 시스템인 `CB-Tumblebug` 과 `CB-Spider` 와 통합하여 기능을 제공하기 때문에 `CM-ANT` 가 올바른 기능을 하기 위해선 동일 시스템 상에서 관련된 서브 시스템이 작동해야한다.

<br/>

# 목    차

1. [실행 환경](#1-실행-환경)
2. [실행 방법](#2-실행-방법)
4. [API 규격](#3-API-규격)
3. [활용 방법](#4-활용-방법)
5. [특이 사항](#5-특이-사항)
6. [활용 정보](#6-활용-정보)
 
***

## 1. 실행 환경

> - OS : Ubuntu 22.04
> - Language : Go 1.21.6
> - Container : Docker 25.0.0
> - SubSystem
>   - CB-Spider : v0.8.17
>   - CB-Tumblebug : v0.8.12

<br/>

## 2. 실행 방법

✨ **필요 패키지/종속성 설치**

```shell
sudo apt update -y
sudo apt install make git
```

<br/>

✨ **Go 설치**

> [Note]<br/>
Install the latest stable version of Go for CM-ANT contribution/development since backward compatibility seems to be supported.

Install Go 1.21.6, see Go all [releases](https://go.dev/dl/) and [Download and install](https://go.dev/doc/install)


```shell
# Set Go version
GO_VERSION=1.21.6

# Get Go archive
wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz

# Remove any previous Go installation and
# Extract the archive into /usr/local/
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz

# Append /usr/local/go/bin to .bashrc
echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc

# Apply the .bashrc changes
source ~/.bashrc

# Verify the installation
echo $GOPATH

# go version go1.21.6 linux/amd64
go version
```

<br/>

✨ **CB-Spider 및 CB-Tumblebug 실행**

CB-Spider와 CB-Tumblebug을 실행 <br/>
(CB-Spider v0.8.17, CB-Tumblebug v0.8.12 로 실행 방식 제공, 필요에 따라 수정).

> 아래는 대략적인 흐름을 나타내며 3개의 터미널이 사용

<br/>

**\>\> 터미널1 실행 (CB-Spder 기동)**

1. CB-Tumblebug 소스코드 다운로드
>참고 - Tumblebug README에서는 `~/go/src/github.com/cloud-barista/cb-tumblebug`를 기본 디렉토리로 활용 - 필요시 변경
```shell
git clone https://github.com/cloud-barista/cb-tumblebug.git $HOME/go/src/github.com/cloud-barista/cb-tumblebug

cd ~/go/src/github.com/cloud-barista/cb-tumblebug

git checkout tags/v0.8.12 -b v0.8.12
```

2. CB-Spider 실행(v0.8.17)
```shell
cd ~/go/src/github.com/cloud-barista/cb-tumblebug

source conf/setup.env

./scripts/runSpider.sh
```

<br/>


**\>\> 터미널2 실행 (CB-Tumblebug 기동)**

3. CB-Tumblebug 소스코드 빌드 및 실행(v0.8.12)
```shell
cd ~/go/src/github.com/cloud-barista/cb-tumblebug

source conf/setup.env

make && make run
```

<br/>


**\>\> 터미널3 실행 (기본 인증 정보 등록)**

4. CSP의 Credential 준비(참고: [CSP별 인증 정보 획득 방법](https://github.com/cloud-barista/cb-tumblebug/wiki/How-to-get-public-cloud-credentials))
<br/>

5. 크리덴셜 파일 생성
파일 생성 방법: 아래 스크립트를 통해 `credentials.yaml` 파일 자동 생성<br/>(경로: ```~/.cloud-barista/```)
  ```shell
  cd ~/go/src/github.com/cloud-barista/cb-tumblebug

  ./scripts/init/genCredencialFile.sh
  ```
<br/>

6. CSP의 Credential 정보로 `~/.cloud-barista/credentials.yaml` 파일 수정
<br/>

7. 멀티 클라우드 연결 정보 및 공통 자원 등록

- CB-Tumblebug을 활용하여 멀티 클라우드 인프라를 생성하기 위해 필요한 자원 사전 등록
  - 클라우드 연결 정보 (CSP, Credential, Region 등)
  - Public Image
  - VM Spec (Instance Type 등) 

```shell
cd ~/go/src/github.com/cloud-barista/cb-tumblebug

./scripts/init/init.sh
```

<br/>

✨ **CM-ANT 실행**
1. CM-ANT 소스코드 다운로드
참고 - 여기에서는 `~/go/src/github.com/cloud-barista/cm-ant`을 기본 디렉토리로 활용 - 필요시 변경

```shell
git clone https://github.com/cloud-barista/cm-ant.git $HOME/go/src/github.com/cloud-barista/cm-ant

cd ~/go/src/github.com/cloud-barista/cm-ant
```
<br/>

2. CM-ANT 실행
```shell
cd ~/go/src/github.com/cloud-barista/cm-ant

make run
```

<br/>

## 3. API 규격

✨ **CM-ANT Swagger Endpoint**
- CM-ANT 서버 기동 후 아래의 엔드포인트에 접근해서 swagger api document 확인 가능
> http://localhost:8880/ant/swagger/index.html


✨ **CM-ANT Swagger Build if Need**
1. swaggo 설치
```shell
go install github.com/swaggo/swag/cmd/swag@latest
```

2. swag build
```shell
cd ~/go/src/github.com/cloud-barista/cm-ant

make swag
```

<br/>

## 4. 활용 방법

- TBD

<br/>

## 5. 특이 사항

- TBD

<br/>

## 6. 활용 정보

- TBD
