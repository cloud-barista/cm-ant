# GitHub Actions Workflows

이 디렉토리는 CM-ANT 프로젝트의 GitHub Actions 워크플로우를 포함합니다.

## 기존 워크플로우

### 1. Continuous Integration (CI)
- **파일**: `continuous-integration.yaml`
- **트리거**: Pull Request 생성/업데이트
- **기능**: 소스 코드 빌드 및 컨테이너 이미지 빌드 테스트

### 2. Continuous Delivery (CD)
- **파일**: `continuous-delivery.yaml`
- **트리거**: 
  - `main` 브랜치 푸시
  - `v*.*.*` 형식의 태그 푸시
- **기능**: Docker Hub 및 GHCR에 컨테이너 이미지 배포

## 새로운 워크플로우

### 3. Rebuild Docker Image
- **파일**: `retag-release.yaml`
- **트리거**: 수동 실행 (`workflow_dispatch`)
- **기능**: 기존 Git 태그 위치는 그대로 유지하고 Docker 이미지만 재빌드

#### 사용법:
1. GitHub 리포지토리의 **Actions** 탭으로 이동
2. **Rebuild Docker Image** 워크플로우 선택
3. **Run workflow** 버튼 클릭
4. 입력값 설정:
   - `tag_name`: 재빌드할 태그명 (예: `v0.4.0`)
   - `confirm_rebuild`: `REBUILD` 입력 (확인용)

#### 특징:
- ✅ **Git 태그 위치 유지**: 기존 커밋 위치 그대로 유지
- ✅ **Docker 이미지만 재빌드**: 같은 코드로 새로운 이미지 생성
- ✅ **안전한 작업**: Git 히스토리에 영향 없음
- ✅ **Digest 충돌 해결**: 기존 Docker 이미지만 삭제

#### 사용 시기:
- Docker Hub에서 같은 태그에 여러 digest가 존재할 때
- Docker 이미지 빌드 과정에서 문제가 있었을 때
- Git 태그 위치는 유지하되 Docker 이미지만 새로 만들고 싶을 때

### 4. Move Tag to Latest Commit ⚠️
- **파일**: `force-rebuild.yaml`
- **트리거**: 수동 실행 (`workflow_dispatch`)
- **기능**: Git 태그를 현재 HEAD 커밋으로 이동하고 Docker 이미지 재빌드

#### 사용법:
1. GitHub 리포지토리의 **Actions** 탭으로 이동
2. **Move Tag to Latest Commit** 워크플로우 선택
3. **Run workflow** 버튼 클릭
4. 입력값 설정:
   - `tag_name`: 이동할 태그명 (예: `v0.4.0`)
   - `confirm_move`: `MOVE_TAG` 입력 (확인용)

#### ⚠️ 주의사항:
- **Git 태그 위치 변경**: 기존 커밋에서 현재 HEAD로 이동
- **기존 참조 깨짐**: 기존에 해당 태그를 참조하던 코드/문서가 깨질 수 있음
- **히스토리 변경**: Git 태그 히스토리가 변경됨
- **복구 불가**: 한번 실행하면 이전 태그 위치로 복구하기 어려움

#### 특징:
- **완전한 재태깅**: Git 태그와 Docker Hub 태그를 모두 삭제 후 재생성
- **최신 코드 반영**: 현재 HEAD 커밋의 모든 변경사항이 반영됨
- **소스-이미지 동기화**: Git 태그 위치와 Docker 이미지 내용이 완벽히 일치

#### 사용 시기:
- 최신 커밋들을 특정 태그에 반영하고 싶을 때
- Git 태그 위치 변경이 허용되는 상황일 때
- 기존 태그 참조를 업데이트할 수 있을 때

## 문제 해결 시나리오

### 시나리오 1: Docker Hub에서 같은 태그에 여러 digest 존재
```bash
# 문제: v0.4.0 태그가 두 개의 다른 digest를 가리킴
# 해결: Rebuild Docker Image 워크플로우 사용
# 결과: Git 태그 위치는 유지, Docker 이미지만 새로 생성
```

### 시나리오 2: 최신 코드를 특정 태그에 반영하고 싶음
```bash
# 문제: 최신 커밋들을 v0.4.0 태그에 반영하고 싶음
# 해결: Move Tag to Latest Commit 워크플로우 사용
# 주의: Git 태그 위치가 변경됨 (기존 참조 깨질 수 있음)
```

### 시나리오 3: 일반적인 새 릴리즈
```bash
# 방법: 기존 CD 워크플로우 사용
git tag v0.4.1
git push origin v0.4.1
```

### 시나리오 4: Docker 이미지 빌드 문제 해결
```bash
# 문제: Docker 이미지 빌드 과정에서 문제 발생
# 해결: Rebuild Docker Image 워크플로우 사용
# 결과: 같은 코드로 새로운 Docker 이미지 생성
```

## 필요한 GitHub Secrets

다음 Secrets이 설정되어 있어야 합니다:

- `DOCKER_USERNAME`: Docker Hub 사용자명
- `DOCKER_PASSWORD`: Docker Hub 비밀번호 또는 액세스 토큰
- `CR_PAT`: GitHub Container Registry Personal Access Token

## 워크플로우 실행 권한

- **cloud-barista** 조직의 멤버만 워크플로우를 실행할 수 있습니다
- 포크된 리포지토리에서는 실행할 수 없습니다 (보안상의 이유)

## 워크플로우 선택 가이드

### 🔄 Rebuild Docker Image vs 🏷️ Move Tag to Latest Commit

| 상황 | 권장 워크플로우 | 이유 |
|------|----------------|------|
| Docker Hub digest 충돌 | **Rebuild Docker Image** | Git 태그 위치 유지, 안전함 |
| Docker 이미지 빌드 문제 | **Rebuild Docker Image** | 같은 코드로 새 이미지 생성 |
| 최신 커밋을 태그에 반영 | **Move Tag to Latest Commit** | Git 태그 위치 변경 필요 |
| 일반적인 새 릴리즈 | **기존 CD 워크플로우** | 새 태그 생성 (v0.4.1) |

### ⚠️ Move Tag to Latest Commit 사용 전 체크리스트

1. **기존 태그 참조 확인**: 문서, 코드, 다른 프로젝트에서 해당 태그를 참조하는지 확인
2. **팀 동의**: Git 태그 위치 변경에 대한 팀원들의 동의
3. **백업**: 필요시 기존 태그 위치 백업
4. **통지**: 관련 팀원들에게 태그 위치 변경 사전 통지

## 트러블슈팅

### 워크플로우가 실행되지 않는 경우
1. GitHub Secrets 설정 확인
2. 리포지토리 권한 확인
3. 입력값 형식 확인 (태그명은 `v0.4.0` 형식)
4. 확인 코드 정확히 입력 (`REBUILD` 또는 `MOVE_TAG`)

### Docker Hub API 오류
1. Docker Hub 계정 권한 확인
2. API 레이트 리미트 확인
3. 네트워크 연결 상태 확인

### Git 태그 관련 오류
1. 기존 태그가 존재하는지 확인
2. 태그 삭제 권한 확인
3. 원격 저장소 접근 권한 확인
