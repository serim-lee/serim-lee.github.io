# 버전과 변경 이력

버전은 정식 `MAJOR.MINOR.PATCH` 형식입니다. [SemVer](https://semver.org/lang/ko/) 기준을 따릅니다.

- **Major**: 기존 형식과 호환되지 않는 변경.
- **Minor**: 기존 형식을 유지하는 기능·선택 항목 추가.
- **Patch**: 형식을 바꾸지 않는 오류·표시 수정.

프로그램 버전은 `app/release.json`, 데이터 묶음의 형식 버전은 `data/version.json`에서
관리합니다. 개별 데이터 JSON에는 버전을 붙이지 않습니다. 개인 정보나 콘텐츠를
수정할 때는 버전을 올리지 않습니다.

릴리스마다 `<버전>.md`에 날짜와 짧은 변경 목록을 적고 `index.json`에 추가합니다.
이미 배포한 노트는 유지하고 다음 버전의 파일을 추가합니다.

`editor/build.bat`은 `app/release.json`의 버전을 `main.compiledVersion`에 넣어
실행 파일을 빌드합니다. 업데이트 후 재시작 시 이 빌드 버전으로 교체 결과를 확인합니다.

`app/release.json`의 `upstream`은 원본 저장소와 조회할 브랜치입니다.
편집기는 이 원본의 최신 버전과 비교하여, 현재 버전보다 높은 노트만 표시합니다.
업데이트 실행 시에는 확인한 버전을 다시 검사하고 원본 커밋 하나를 고정해 프로그램
파일을 내려받습니다. 기존 프로그램을 백업·적용한 뒤 편집기 종료를 기다리는 외부
PowerShell 프로세스가 실행 파일을 교체합니다. 다음 실행에서 미완료 작업을 검사·복구하고
적용·되돌림·실패 결과를 표시합니다.

업데이트 교체·보존·재생성 경계는 `app/update-policy.json`에 기록합니다.
`data/`, `exports/`, `CNAME`, `.git/`, 개인 작업 설정과 `editor/runtime/`는 보존합니다.
개인 데이터 버전, Git의 수정 상태, 기존 코드 해시 일치는 업데이트 사전 검사에서 제외합니다.

검사 명령은 저장소 루트에서 실행합니다.

```powershell
.\app\tools\check-release.bat
.\app\tools\check-release.bat --preserve-data
```

기본 검사는 프로그램·데이터 버전, 릴리스 노트와 업데이트 경계를 확인합니다.
`--preserve-data`는 개인 데이터의 형식·버전 검사를 생략하는 배포·업데이트 검사 모드입니다.
