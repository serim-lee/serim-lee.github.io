# Research Portfolio Site Kit

연구자 한 사람의 구조화된 데이터를 반응형 영문·국문 홈페이지와 CV로 보여 주는 정적 사이트 도구입니다. 이 저장소를 fork한 뒤 `data/`만 자신의 정보로 교체하면 GitHub Pages에서 별도 서버나 데이터베이스 없이 사용할 수 있습니다.

현재 들어 있는 데이터는 예시가 아니라 실제 개인 정보입니다. 공개하기 전에 `data/`의 프로필, 연락처, 논문, 프로젝트, 이미지 등을 모두 자신의 내용으로 확인하거나 교체하세요.

## 빠르게 시작하기

1. 이 저장소를 fork합니다.
2. 개인 홈페이지로 사용할 경우 저장소 이름을 `<GitHub 사용자명>.github.io`로 지정합니다. 일반 저장소 이름을 사용하면 프로젝트 사이트로도 배포할 수 있습니다.
3. `data/` 안의 파일과 `data/media/`의 미디어를 자신의 내용으로 교체합니다.
4. Windows에서 `app/tools/serve.bat`을 실행해 로컬 미리보기를 엽니다.
5. `app/tools/validate.bat`을 실행해 형식, ID 참조, 파일 경로를 검사합니다.
6. push 전에 `app/tools/build.bat`으로 내보내기 파일과 데이터 보고서를 갱신합니다.
7. 변경 사항을 push합니다.
8. GitHub 저장소의 **Settings → Pages**에서 **Deploy from a branch**, 배포 브랜치, `/(root)`를 선택합니다. 자세한 절차는 [GitHub Pages 배포 소스 안내](https://docs.github.com/en/pages/getting-started-with-github-pages/configuring-a-publishing-source-for-your-github-pages-site)를 참고하세요.

`data/*.json`이 홈페이지와 CV가 사용하는 원본 데이터입니다. 배포된 화면은 이
파일들을 직접 읽으므로 콘텐츠 변경을 홈페이지와 CV에 반영하기 위한 별도 빌드는
필요하지 않습니다. `exports/` 아래의 BibTeX·CSL 내보내기와 데이터 보고서는
파생 파일이므로 이를 갱신할 때만 `app/tools/build.bat`을 실행합니다.

## 지원 기능

- 영문·국문 전환과 언어별 fallback
- 프로필, 경력, 학력, 장학, 자격사항, 수상, 교육, 기술 역량과 학술활동
- 프로젝트와 소프트웨어 카드 및 이미지·동영상 갤러리
- 논문 검색, 유형·주제·공저자 필터와 최신순·인용순·유형별·주제별 보기
- 인용수, 출처·조회일 미리보기와 인용 논문 목록 링크
- 본인 저자 객체와 저자 순서를 이용한 주저자 자동 판별
- 저자별 소속 등의 마우스오버 정보
- under review, in press, 초록, 주제어, 비고 표시
- BibTeX 복사·내보내기와 전체 서지 내보내기 파일 생성
- 스크롤형 기록 페이지, 미디어 캐러셀, 확대 보기, 키보드·스와이프 조작
- 같은 데이터를 사용하는 영문·국문 CV와 브라우저 PDF 저장
- 데이터를 보존하는 웹 탭 표시 설정과 CV 섹션·항목 선택
- 모바일 탭 간 스와이프 이동
- 프레임워크와 외부 데이터베이스가 필요 없는 GitHub Pages 배포

## 저장소 구조와 편집 경계

| 위치 | 역할 | 일반 사용자의 편집 여부 |
|---|---|---|
| `data/` | 프로필과 콘텐츠 원본, 설정, 이미지·동영상, 공통 데이터 버전 | **편집하는 영역** |
| `exports/` | 인용수 조회 결과, 서지 내보내기, 데이터 보고서 | 도구로 갱신 |
| `app/` | 화면 코드, 스타일, 아이콘, 검증 도구, 릴리스 정보 | 기능을 개발할 때만 편집 |
| `editor/`, `Profile-Editor.exe` | 편집기 코드와 실행 파일 | 프로그램 업데이트 대상 |
| `index.html`, `cv.html` | GitHub Pages 진입점 | 일반적으로 편집하지 않음 |
| `.github/` | push·PR 검증 자동화 | 일반적으로 편집하지 않음 |
| `.gitignore`, `.gitattributes`, `.nojekyll` | Git 및 GitHub Pages 필수 설정 | 루트에 유지 |

콘텐츠만 바꾸려는 사용자는 `data/` 이외의 구조를 알 필요가 없습니다.
`data/*.json`은 편집기의 임시 복사본이 아니라 항상 유효한 원본이므로 Profile-Editor가
닫혀 있을 때 직접 수정해도 다음 실행 및 홈페이지·CV 로딩에 그대로 반영됩니다.
디자인이나 동작을 수정하려는 개발자는 `app/assets/`, 검사·내보내기 기능을
수정하려는 개발자는 `app/tools/`를 사용합니다.

## 버전과 업데이트 경계

프로그램 버전은 `app/release.json`, 데이터 묶음의 공통 형식 버전은
`data/version.json`에서 관리합니다. 현재 두 버전은 `1.0.0`입니다.
개별 데이터 JSON은 기존 배열·객체 구조를 유지합니다. 콘텐츠를 편집할 때는
버전을 바꾸지 않습니다. 과거 `schema_version`이 있는 파일도 읽을 수 있습니다.

Major는 호환되지 않는 변경, Minor는 호환되는 기능 추가, Patch는 오류·표시 수정입니다.
변경 이력과 작성 기준은 [app/releases/](app/releases/README.md)에 모읍니다.
`index.json`은 릴리스 버전·날짜·노트 경로를 최신순으로 나열합니다.

업데이트 경계는 `app/update-policy.json`에 기록합니다.

| 구분 | 처리 |
|---|---|
| `app/`, `editor/`, 루트 HTML·실행 파일 등 명시된 프로그램 파일 | 교체 대상 |
| `data/`, `exports/`, `CNAME`, `.git/`, 개인 작업 설정, `editor/runtime/` | 보존 |
| `exports/publications.bib`, `.ris`, `.csl.json`, `build-report.json` | `app/tools/build.bat`의 재생성 대상 |
| 목록에 없는 파일 | 보존 |

프로그램 버전과 데이터 버전·릴리스 노트의 연결, 업데이트 경계는
`app/tools/check-release.bat`으로 검사합니다. `--preserve-data`를 붙이면 개인 데이터의
형식·버전 검사를 생략하고 프로그램 릴리스·노트·업데이트 경계를 검사합니다.
데이터 형식 이전은 업데이트와 별도로 처리합니다.

편집기의 **설정 → 업데이트 → 업데이트 확인**은 현재 저장소와 원본 저장소의 프로그램
버전을 비교합니다. 현재 버전보다 높고 원본 최신 버전 이하인 릴리스 노트만 버전별로
표시하며, 새 버전이 있으면 **업데이트 실행** 버튼을 활성화합니다. 원본 저장소와 조회
브랜치는 `app/release.json`의 `upstream`에 기록하므로 fork의 `origin`과 별도로 유지됩니다.

**업데이트 실행**을 누르면 편집 중인 변경을 먼저 저장합니다. 저장에 실패하면
업데이트도 중단합니다. 원본 브랜치의 커밋 하나를 고정해 프로그램 파일을 내려받고,
기존 프로그램 파일을 백업한 뒤 적용합니다. 편집기가 종료되면 외부 PowerShell
프로세스가 `Profile-Editor.exe`를 교체합니다. 다음 실행에서는 미완료 작업과 실행
파일의 버전을 확인하고, 적용·되돌림·실패 결과를 **설정 → 업데이트**에 표시합니다.

개인 데이터 버전, Git의 수정 상태, 기존 코드 해시의 일치는 업데이트 사전 검사에서
제외합니다. 프로그램 파일의 로컬 수정은 원본으로 교체되며, 교체 전 파일은
`editor/runtime/updates/`에 백업합니다.

## 데이터 파일

| 파일 | 내용 |
|---|---|
| `data/settings.json` | 홈페이지·CV 색상과 구성, 프로젝트·논문 주제, 학회·저널 목록 |
| `data/profile.json` | 이름, 직책, 소속, 소개, 연락처, 경력, 학력, 특허, 장학, 자격, 교육, 기술 |
| `data/academic_activities.json` | 편집·전문·학술대회 서비스, 초청 발표, 리뷰 기록 |
| `data/people.json` | 논문 저자·특허 발명자, 본인 표시, 사람별 보조 정보 |
| `data/publications.json` | 논문·학술대회 발표와 초록, 주제어, 상태, 링크 |
| `data/awards.json` | 수상명, 수상일, 수여기관으로 구성된 수상 기록 |
| `data/projects.json` | 프로젝트 설명과 미디어 갤러리 |
| `data/software.json` | 소프트웨어 설명, 기술 아이콘, 미디어 갤러리, 저장소·배포·웹 링크 |
| `data/board.json` | 제목, 본문, 날짜, 여러 미디어와 캡션으로 구성된 게시판 |
| `data/media/` | 사용자가 관리하는 프로필·프로젝트·소프트웨어·기록 이미지와 동영상 |
| `data/version.json` | 전체 원본 데이터가 공유하는 형식 버전 |

작업 결과는 `exports/`에 저장합니다. `publication-citations.json`은 DOI별 인용수와
조회일·출처 식별자이며, `publications.bib`, `publications.ris`, `publications.csl.json`,
`build-report.json`은 내보내기·빌드 결과입니다. CSL JSON은 외부 표준 배열 형식을 유지합니다.

## 학회·저널 목록

**주제** 탭은 논문과 프로젝트가 함께 쓰는 공통 주제 목록을 관리합니다.
`settings.json`의 `topics`에 이름·색상·순서를, `topic_fallback`에 기타 분류 이름을
저장합니다. 논문의 `topic`과 프로젝트의 `theme`는 같은 ID를 참조하고,
사용 건수와 삭제 제한도 합산합니다. 이전의 분리된 목록은 불러올 때 ID를 유지해
합치고, 편집기에서 저장할 때 새 형식으로 기록합니다.

편집기에서 **주제 → 학회 → 저널** 순서로 표시되는 탭에서 독립 목록을
편집합니다. 상단 **모두 저장**을 누르면 `settings.json`의 `conferences`, `journals`
배열에 저장하며, **변경 되돌리기**로 저장 전 상태를 복원할 수 있습니다.
논문 유형이 국제·국내 학술지이면 저널 목록에서, 국제·국내 학술대회이면 학회 목록에서
항목을 선택할 수 있습니다. 선택한 ID는 `publications.json`의 선택 속성 `journal_id` 또는
`conference_id`에 저장합니다. 심사 중인 논문처럼 게재처가 정해지지 않은 항목은 비워 둡니다.
수동 게재지 입력 대신 등록된 저널·학술대회와 논문별 서지 속성으로 참고문헌을 자동 구성합니다.

각 항목의 고유 `id`는 편집기에서 추가할 때 자동 생성되며, 목록과 입력 폼에는 표시하지 않습니다.
기존 ID와 논문 연결은 유지되고, 이름을 바꿔도 ID는 바뀌지 않습니다.
영문 또는 국문 이름을 입력해야 하며, 나머지 값은 비워 둘 수 있습니다.
저널 ID는 소문자 영문·숫자·단일 하이픈을 사용합니다. 학술대회 ID는 `ASIM 2024`,
`SASBE 2024`, `AIK Conf. 2024F`, `AIK Conf. 2024S`, `KIAEBS 2025`처럼 영문·숫자·
공백·하이픈·약어 마침표를 사용할 수 있습니다. `F`는 추계, `S`는 춘계를 구분하며,
기존 소문자 ID도 지원합니다. ID는 같은 목록 안에서 중복될 수 없습니다.
저널 이름·학술대회명(`name`), 저널 출판사(`publisher`), Proceedings 발행기관(`proceedings_publisher`),
장소(`location`), Proceedings 이름(`proceedings_title`), 시리즈명(`proceedings_series`)은
각각 `_en`·`_ko` 속성으로 영문·국문을 저장합니다. 이전의 단일 문자열은 불러올 때
한글이 포함되면 국문, 그 외에는 영문으로 옮기며, 편집기에서 저장할 때 새 형식으로 기록합니다.
등재 구분 `indexing`은 `없음`, `SCIE`, `SCOPUS`, `KCI` 중 하나만 선택하며,
`없음`은 빈 배열로 저장합니다.
학회 편집기의 **Proceedings 정보**는 이름·발행기관·권·호·시리즈명·시리즈 권 순서로
입력합니다. `proceedings_volume`, `proceedings_issue`, `proceedings_series_volume`은
선택 입력 문자열이며, 숫자와 문자를 사용할 수 있습니다.
행사 날짜는 `YYYY-MM-DD`, Proceedings 발행일은 `YYYY`, `YYYY-MM`, `YYYY-MM-DD`를
지원합니다. ISSN은 `1234-567X`, URL은 http 또는 https 형식입니다.
학회 종료일을 입력할 때는 시작일도 입력해야 합니다.
논문에 연결된 항목은 선택을 해제한 뒤 삭제할 수 있습니다.

```json
{
  "journals": [
    {
      "id": "example-journal",
      "name_en": "Example Journal",
      "name_ko": "예시 저널",
      "publisher_en": "Example Publisher",
      "publisher_ko": "예시 출판사",
      "issn": "1234-567X",
      "url": "https://example.org/journal",
      "indexing": ["SCIE"]
    }
  ],
  "conferences": [
    {
      "id": "CONF 2026",
      "name_en": "Example Conference 2026",
      "name_ko": "예시 학술대회 2026",
      "start_date": "2026-10-18",
      "end_date": "2026-10-20",
      "location_en": "Seoul, Korea",
      "location_ko": "대한민국 서울",
      "url": "https://example.org/conference",
      "proceedings_publisher_en": "Example Publisher",
      "proceedings_publisher_ko": "예시 발행기관",
      "proceedings_title_en": "Proceedings of Example Conference 2026",
      "proceedings_title_ko": "예시 학술대회 2026 논문집",
      "proceedings_volume": "2",
      "proceedings_issue": "1",
      "proceedings_series_en": "Example Research Series",
      "proceedings_series_ko": "예시 연구 시리즈",
      "proceedings_series_volume": "42",
      "proceedings_date": "2026-11-01",
      "proceedings_issn": "2345-678X",
      "indexing": ["SCOPUS"]
    }
  ]
}
```

## 논문 서지 정보 입력

논문 탭의 **서지 정보**에서 학술지 논문은 권(`volume`), 호(`issue`), 쪽수(`pages`),
논문번호(`elocator`), 온라인 게재일(`online_date`), Issue 발행일(`issue_date`)을 입력합니다.
학술대회 논문은 쪽수(`pages`)를 입력합니다. 모두 선택 문자열 속성이므로 이전 논문에는
없어도 되며, 권·호·쪽수·논문번호에 문자나 쪽 범위를 사용할 수 있습니다.
날짜는 `YYYY`, `YYYY-MM`, `YYYY-MM-DD` 형식이며 입력한 정밀도를 유지합니다.
논문 탭에서 국제·국내 학술지 논문을 펼치면 주제어 아래에 ‘온라인 게재’와 ‘권호 발행’ 날짜를
표시하고, 그 뒤에 저널에 등록된 ‘출판사’를 같은 줄로 이어서 표시합니다. 입력된 정보만 표시하며,
출판사는 화면 언어에 맞는 국문·영문 이름을 사용합니다. 초록·주제어가 없어도 게재일이나
출판사 정보가 있으면 펼칠 수 있습니다.
국제·국내 학술대회 논문을 펼치면 같은 위치에 ‘학술대회’ 정보로 학회명·행사 기간·장소를
표시합니다. 그 아래 줄에는 Proceedings의 ‘권호 발행’ 날짜(`proceedings_date`)와
‘출판사’(`proceedings_publisher_en`·`proceedings_publisher_ko`)를 이어서 표시합니다.
입력된 정보만 표시하며, 출판사는 화면 언어에 맞는 이름을 사용합니다.
초록·주제어가 없어도 학술대회 또는 Proceedings 발행 정보가 있으면 펼칠 수 있습니다.
학술지의 출판 날짜는 `issue_date` → `online_date`, 학술대회는 등록된
`proceedings_date` → `end_date` 순으로 결정합니다. 이 날짜를 홈페이지·CV·정렬·
타임라인·인용 연도와 서지 내보내기에 함께 사용합니다. 편집기의 서지 정보 아래에서
날짜와 참고문헌 미리보기를 확인할 수 있습니다. 저널은 이름·권·호·쪽수 또는 논문번호,
학술대회는 Proceedings·시리즈·쪽수·행사 기간·장소를 현재 참고문헌 양식으로 표시합니다.
논문 유형 옆에는 등록된 저널·학술대회의 등재 구분(`SCIE`, `SCOPUS`, `KCI`)을
로고 없는 문자 태그로 표시합니다. 학술대회는 Proceedings·시리즈·쪽수를 참고문헌 줄에,
학술대회명·행사 기간·장소를 펼친 상세 정보의 주제어 아래에 표시합니다. 홈페이지 수상의
연결 논문은 Proceedings 서지 정보와 학술대회 정보를 두 줄로 표시합니다.
CV의 국내·국외 학술대회 논문과 수상의 연결 논문에는
Proceedings 서지 정보를 표시하며, 학술대회명·행사 기간·장소 줄은 표시하지 않습니다.
인용 복사와 서지 파일에는 두 줄의 내용을 함께 포함합니다.
국제학술지·국제학술대회의 인용 복사와 복사 미리보기는 화면 언어에 관계없이
저자명·논문 제목·저널명·학술대회명 등 영문 정보를 사용합니다.
논문의 **색인 / Index** 필터는 전체·SCIE·SCOPUS·KCI·기타를 지원합니다.
각 색인은 등록된 등재 구분을 기준으로 하며, 기타는 색인 정보가 없는 항목을 선택합니다.
검색·유형·주제·공저자·주저자 필터와 함께 적용되며, 목록·건수·타임라인이 같이 갱신됩니다.

## 논문 인용수

페이지를 먼저 표시한 뒤 인용수를 백그라운드에서 불러옵니다. OpenAlex에 등재된
발행 논문은 DOI를 묶어 키 없는 무료 API로 조회하며, 브라우저에 결과를 하루 동안
저장합니다. 조회 실패 시 기존 수치와 실제 조회일을 유지합니다. 조회되지 않은
논문은 0회로 표시하지 않습니다. 인용순에서는 확인된 수치를 내림차순으로,
동률과 미조회 논문을 최신순으로 정렬합니다.

조회 결과는 `exports/publication-citations.json`에 저장합니다. 아래 명령으로
Go 네이티브 도구로 OpenAlex 수치를 갱신할 수 있습니다. API 키는 필요하지 않습니다.

```powershell
app/tools/refresh-citations.bat
```

실행 결과에는 인용수가 실제로 달라진 논문만 `논문명: cite 이전 -> 이후`로
표시합니다. 처음 조회된 논문은 `not previously recorded -> 인용수`로 구분하고,
조회·변경·신규 기록 건수를 각각 집계합니다. 변화가 없으면
`No citation count changes.`를 출력하며, 정상 조회된 논문의 조회일은 갱신합니다.

국내 논문은 [KCI 논문 상세](https://www.kci.go.kr/kciportal/main.kci)의 인용현황도
확인할 수 있습니다. KCI 검색 OpenAPI는 인증키를 요구하고 공개 서비스의 직접
접근에는 제한이 있어, 국내 논문은 공개 페이지에서 확인한 KCI 수치와
조회일을 저장해 표시합니다. KCI 수치는 OpenAlex 수치와 합산하지 않습니다.
클릭하면 해당 논문의 KCI 인용현황으로 이동합니다. KCI가 인용논문 목록을 아직
공개하지 않은 경우 팝업에 공개 대기 상태를 표시합니다.

KCI에서 새 수치를 확인한 후 아래 명령으로 그 논문의 수치와 조회일을 갱신합니다.
조회 기록과 출처 식별자는 `exports/publication-citations.json`에 저장됩니다.

```powershell
app/tools/refresh-citations.bat --kci-only --kci-count "<DOI>" 1
```

## VS Code 스니펫

VS Code에서 JSON 배열 안에 아래 접두어를 입력하고 스니펫을 선택하면 필요한 필드가 생성됩니다. 이후 `Tab`을 눌러 빈 값을 차례대로 입력합니다.

| 접두어 | 생성 항목 |
|---|---|
| `project` | `projects.json` 프로젝트 |
| `software` | `software.json` 소프트웨어 |
| `publication` | `publications.json` 논문·학술대회 발표 |
| `award` | `awards.json` 수상 |
| `board` | `board.json` 게시물 |
| `media` | 프로젝트·소프트웨어·게시판 미디어 |

스니펫은 필드 구조만 만들며 날짜, ID, 설명, 링크 등의 내용은 자동으로 채우지 않습니다. 여러 항목 사이의 쉼표는 JSON 문법에 맞게 직접 입력합니다.

## 공통 데이터 규칙

- 사람, 수상, 소프트웨어, 프로젝트 테마, 논문 주제의 `id`는 JSON 문자열이어야
  하며 정규식 `^[a-z0-9]+(?:-[a-z0-9]+)*$`와 정확히 일치해야 합니다. 즉 영문
  소문자·숫자와 단어 사이의 하이픈만 사용하며 숫자형 값, 대문자, 밑줄, 앞뒤
  공백은 허용하지 않습니다. 각 ID는 해당 목록 안에서 고유해야 합니다.
- `author_ids`, 비어 있지 않은 `award_id`, 프로젝트의 `theme`, 비어 있지 않은
  논문의 `topic`도 같은 ID 형식이어야 하며 대상 목록에 실제로 존재해야 합니다.
- ID를 직접 바꿀 수는 있지만 반드시 Profile-Editor를 닫은 상태에서 해당 ID를
  참조하는 모든 필드를 함께 변경해야 합니다. 일부 참조만 남기면 검증에
  실패합니다.
- 프로젝트, 논문과 학술활동 항목에는 의도적으로 자체 `id` 필드를 두지 않습니다.
- `_en`은 영문, `_ko`와 `_kr`은 국문 값입니다. 선택한 언어가 비어 있으면 가능한 공식 표기로 fallback합니다.
- 여러 ID나 키워드는 JSON 문자열 배열(`[]`)로 입력합니다.
- Boolean 값은 따옴표 없는 `true` 또는 `false`를 사용합니다.

현재 관계 필드는 다음과 같습니다.

- `projects.json`의 `theme` → `settings.json`의 `topics[].id`
- `publications.json`의 `topic` → `settings.json`의 `topics[].id`
- `publications.json`의 `author_ids` → `people.json`
- `publications.json`의 `award_id` → `awards.json`

검증 도구는 중복 ID, 존재하지 않는 참조, 잘못된 날짜·상태·파일 경로를 거부합니다.

메인 페이지 오른쪽 영역의 노출 항목과 순서는 `settings.json`의 `main_page_sections` 배열에서 관리합니다. 숨길 항목은 `hidden_main_page_sections`로 옮깁니다. 두 배열에는 다음 키가 중복 없이 각각 한 번씩 포함되어야 합니다.

```json
"main_page_sections": [
  "experience",
  "education",
  "scholarships",
  "certifications",
  "awards",
  "academic_activities",
  "patents",
  "teaching",
  "skills"
],
"hidden_main_page_sections": []
```

메인 페이지와 CV에서 학술활동과 특허를 지원하는 `settings.json`은 공통 데이터 버전
`data/version.json`을 따릅니다. 메인 페이지 및 CV의 노출/숨김 배열은 각각 합쳤을 때
`academic_activities`, `patents`를 포함한 지원 키가 정확히 한 번씩 있어야 합니다. 항목 누락,
중복 또는 정의되지 않은 키가 있으면 데이터 검증과 페이지 로딩이 실패합니다.

## 프로필 관리

### 특허

Profile-Editor의 **프로필 → 특허**에서 `+`로 추가합니다. 원본 데이터는
`profile.json`의 `patents` 배열입니다. 홈페이지와 CV는 **국내특허 [출원]**,
**국내특허 [등록]**처럼 종류·상태별로 묶고 각 구분 안에서 입력 순서를 사용합니다.
기본 위치는 학술활동 다음, 기술 역량 이전입니다. 항목이 비어 있으면 공개 화면에서
숨기고 설정된 위치는 유지합니다. **CV 구성**에서 개별 특허의 포함 여부를 정합니다.

| 필드 | 내용 |
|---|---|
| `title_en`, `title_ko` | 영문·국문 제목. 한 언어 이상 입력 |
| `type` | `domestic` 국내특허, `international` 국제특허 |
| `application_number`, `application_date` | 출원번호·출원일 |
| `publication_number`, `publication_date` | 공개번호·공개일 |
| `registration_number`, `registration_date` | 등록번호·등록일 |
| `inventor_ids` | `people.json`의 사람 ID 배열. 선택한 발명자 순서대로 저장 |
| `applicant_en`, `applicant_ko` | 영문·국문 출원인 |
| `url` | 특허 문헌의 HTTP(S) 링크 |
| `visible_in_CV` | CV 포함 여부. 생략하면 포함 |

번호는 국가별 표기를 보존하는 문자열입니다. 출원·공개·등록번호 중 하나 이상을
입력합니다. 상태는 등록번호가 있으면 **등록**, 공개번호가 있으면 **공개**, 출원번호만
있으면 **출원**으로 표시합니다. 상태는 별도로 저장하지 않습니다.
날짜는 선택 입력이며 `YYYY-MM-DD`
형식입니다. 출원일을 함께 입력하면 공개일·등록일은 출원일과 같거나 늦어야 합니다.
발명자는 논문 저자와 같은 **사람** 목록에서 고르고 순서를 조정합니다. 사람의 이름을
수정하면 논문과 특허에 함께 반영됩니다. 출원번호·공개번호·등록번호의 구분은
[WIPO 특허정보 안내](https://www.wipo.int/edocs/pubdocs/en/wipo-pub-rn2021-1e-en-wipo-guide-to-using-patent-information.pdf)를 참고하세요.

이전 형식의 설정에는 특허를 숨긴 섹션으로 추가하여 기존 표시 순서를
유지합니다. 해당 저장소에서 사용하려면 프로필의 **페이지 구성**과 **CV 구성**에서
특허를 노출합니다.

### 색상과 기본 정보

Profile-Editor의 **설정** 탭에서 에디터·홈페이지·CV에 적용할 색상 테마를 선택합니다.
**모두 저장**을 누르면 `settings.json`의 `color_theme`에 저장됩니다. 지원하는 값은
`sage`(기본), `ocean`, `violet`, `rose`, `sand`, `slate`, `mint`, `teal`, `sky`,
`indigo`, `amber`, `graphite` 및 `custom-1`, `custom-2`, `custom-3`이며, 필드를 생략하면 `sage`를 사용합니다.
선택한 테마의 용도별 색상 코드는 오른쪽 패널에 표시합니다. 사용자 지정 테마 3개는
이름과 색상을 편집할 수 있고, 기본 테마의 색상에서 새로 시작할 수 있습니다.
에디터에서는 저장 전에도 미리 적용하며, **변경 되돌리기**로 저장된 색상을 복원합니다.
사용자 지정 테마는 선택한 기본 테마의 색상을 상속하고 `colors`에 입력한 HEX 코드만 덮어씁니다.

```json
"color_theme": "custom-1",
"custom_themes": {
  "custom-1": {
    "name": "내 테마",
    "base_theme": "ocean",
    "colors": { "accent": "#315f7d", "bg": "#f5f8fb" }
  }
}
```

슬롯은 `custom-1`~`custom-3`이며, 이름은 1~80자, `base_theme`는 기본 테마 12개 중 하나입니다.
색상은 배경·글자·선·버튼·상태 알림 등 26개 속성에 `#RRGGBB` 형식으로 입력합니다.
수정하지 않은 색상은 기본 테마를 사용하고, 저장된 사용자 테마가 없는 슬롯은 세이지로 시작합니다.

`profile.json` 하나에서 메인 화면과 CV의 개인 정보를 함께 관리합니다. 프로필 카드의 대표 미디어는 `media` 객체로 입력합니다.

```json
"media": {
  "src": "profile.jpg",
  "type": "image"
}
```

동영상에는 선택적으로 포스터 이미지를 지정할 수 있습니다.

```json
"media": {
  "src": "profile.mp4",
  "type": "video",
  "poster": "profile.jpg"
}
```

경력, 학력, 장학, 자격사항처럼 반복되는 항목은 배열에 객체를 추가합니다. 장학에는
자유형 `details`를 사용하지 않고 금액을 다음 `amount` 객체로 저장합니다.

```json
"amount": {
  "value": 5000000,
  "currency": "KRW"
}
```

`value`는 JSON 숫자이고 `currency`는 `KRW`, `USD`처럼 대문자 ISO 3자리 통화
코드입니다. 메인 페이지와 CV는 국문에서 KRW, 영문에서 USD로 금액을 표시합니다.
원통화가 표시 통화와 같으면 외부 API를 호출하지 않습니다. 서로 다를 때만 실제 표시가
필요한 시점에 Frankfurter v2 rate endpoint에서 환율을 한 번 가져와 캐시하며, 조회에
실패하면 환산하지 않은 원금과 원통화를 표시합니다. KRW와 USD 금액은 모두 일의
자리에서 반올림해 끝자리가 0인 십 단위로 표시합니다. Profile-Editor에서는 장학의
금액과 통화 코드를 수정할 수 있으며, 다른 변경과 마찬가지로 상단 **저장** 버튼을
눌렀을 때만 `profile.json`에 반영됩니다.

학력(`education`)에는 선택적으로 `"gpa": {"value": 3.0, "max": 4.0}`을 추가할 수
있습니다. `value`는 획득점수, `max`는 만점이며 둘 다 JSON 숫자입니다. 메인 페이지와
CV는 졸업논문명 아래에 `GPA: 3.0 / 4.0`을 한 줄로 표시합니다. 논문명이 없는 학력은
나머지 학력 정보 아래에 표시합니다. `gpa`를 생략하면 표시하지 않습니다.
Profile-Editor의 학력 항목에서 두 점수를 입력하고, 모두 비우면 GPA를 제거할 수 있습니다.

학력에는 선택적으로 `"thesis_url": "https://example.org/thesis"`를 추가할 수 있습니다.
국문·영문이 같은 링크를 사용하며, 메인 페이지와 CV에서 학위논문명에 밑줄과 링크를
표시합니다. 속성을 생략하거나 빈 문자열로 두면 기존처럼 논문명만 표시합니다.
Profile-Editor의 **학위논문 링크 (선택)** 입력란에서 http 또는 https 주소를 수정합니다.

## 수상·자격사항 PDF

수상(`awards.json`)과 자격사항(`profile.json`의 `certifications`)에는 선택적으로
`"pdf": "수상-202011KIAEBS_1.pdf"`처럼 PDF 파일명을 입력할 수 있습니다. 파일은
`data/document`에 모으며, 파일명에는 폴더 경로나 URL을 넣지 않습니다. `pdf`를
생략하거나 빈 문자열로 두면 버튼을 표시하지 않습니다. 연결된 항목은 메인 페이지
우상단의 `pdf` 버튼으로 새 탭에서 PDF를 엽니다. Profile-Editor에서도 각 항목의
PDF 영역에 파일 한 개를 드롭해 추가·교체하고 연결을 제거할 수 있습니다. 드롭한
파일은 **모두 저장**을 누를 때만 `data/document`와 JSON에 반영됩니다. 같은 이름의
파일이 있으면 새 파일명에 접미사를 붙여 기존 파일을 보존합니다.

## 학술활동 관리


`academic_activities.json`은 공통 데이터 버전을 따르는 JSON 객체이며 다음 다섯 배열을
가집니다. 각 항목은 다른 데이터와 연결되지 않는 독립 기록이므로 `id`를 저장하지
않습니다. 영문·국문 필드는 두 값 가운데 하나 이상을 입력하며, 선택 언어의 값이
비어 있으면 다른 언어 값으로 fallback합니다.

| 분류 | 필드 |
|---|---|
| `editorial_service` | `journal_en`, `journal_ko`, `role_en`, `role_ko`, `start_date`, `end_date` |
| `professional_service` | `organization_en`, `organization_ko`, `role_en`, `role_ko`, `start_date`, `end_date` |
| `conference_service` | `conference_en`, `conference_ko`, `role_en`, `role_ko`, `start_date`, `end_date` |
| `invited_talks` | `event_en`, `event_ko`, `topic_en`, `topic_ko`, `date` |
| `reviews` | `journal_en`, `journal_ko`, `YYYY-MM-DD` 문자열 배열 `completed_dates` |

모든 날짜는 `YYYY-MM-DD` 형식입니다. 리뷰의 `completed_dates`에는 완료한 리뷰의
날짜를 하나씩 저장해 실제 완료일을 기억할 수 있습니다. 홈페이지에 표시하는 리뷰
횟수는 이 배열의 길이로 자동 계산하므로 숫자형 `review_count`는 직접 입력하거나
저장하지 않습니다. 학술대회 서비스의 `end_date`가 비어 있으면
시작일을 단일 날짜로 표시하고, 값이 있으면 기간으로 표시합니다. 전문·편집 서비스의
`end_date`는 진행 중이면 빈 문자열로 둘 수 있습니다. 빈 배열에
해당하는 세부분류는 홈페이지에 표시되지 않으며, 다섯 배열이 모두 비어 있으면
`main_page_sections` 설정과 관계없이 학술활동 영역 전체가 자동으로 숨겨집니다.
Profile-Editor의 **학술활동** 탭은 다섯 분류를 각각 별도 목록으로 표시합니다. 각
목록에서 항목을 추가·수정·삭제하고 위·아래 버튼으로 순서를 바꿀 수 있으며, 이 순서는
해당 JSON 배열과 홈페이지·CV 표시 순서에 그대로 반영됩니다. 각 항목의
`visible_in_CV`는 CV 포함 여부이며 생략하면 `true`로 처리됩니다. CV에서는 선택된
항목만 같은 다섯 분류로 나누어 표시하고, 선택 항목이 없는 분류와 학술활동 섹션은
자동으로 숨깁니다. 모든 변경 내용은 상단 **저장** 버튼을 눌렀을 때만 JSON에
반영됩니다.

## 논문과 저자 관리

논문을 추가할 때는 다음 순서를 권장합니다.

1. 새 저자가 있으면 `people.json`에 고유 `id`와 이름을 추가합니다.
2. `publications.json` 배열에 논문 객체를 추가합니다.
3. `author_ids`에 논문상의 저자 순서대로 사람 ID를 입력합니다.
4. `app/tools/validate.bat`을 실행합니다.

중요한 논문 필드는 다음과 같습니다.

- `title_en`, `title_ko`: 출판사·학회 등 공식 출처에 실린 제목. 공식 번역이 없으면 직접 번역하지 않고 빈칸으로 둡니다.
- `abstract_en`, `abstract_ko`: 공식 출처에 공개된 해당 언어의 초록. 요약·번역·추론해서 만들지 않으며, 공식 초록이 확인되지 않으면 빈칸으로 둡니다.
- `keywords_en`, `keywords_ko`: 공식 출처의 주제어를 소문자로 작성한 문자열 배열. 주제어를 임의로 추가하거나 번역하지 않습니다.
- `author_ids`: `people.json`의 ID를 저자 순서대로 나열한 문자열 배열.
- `publication_type`: `international-journal`, `domestic-journal`, `international-conference`, `domestic-conference` 중 하나.
- `date`: 기존 소비자를 위한 파생 날짜 캐시. 편집기에서 논문을 저장할 때 위 우선순위로 갱신합니다. 표시·정렬·내보내기는 이 캐시를 직접 신뢰하지 않고 등록 정보에서 다시 계산하므로 수동으로 수정하지 않습니다.
- `under_review`, `in_press`: JSON Boolean이며 둘 중 하나만 `true`로 설정할 수 있습니다.
- `topic`: `settings.json`에 등록된 주제 ID. 빈 문자열은 기타로 분류되며, 비어
  있지 않은 값은 반드시 등록된 ID여야 합니다.
- `journal_id`, `conference_id`: 등록된 저널 또는 학술대회의 ID. 수동 `venue`, `venue_en`, `venue_kr` 속성은 사용하지 않습니다. 국내 논문은 국문·영문 이름을 선택 언어에 맞게 사용하고, 국제 논문은 영문 참고문헌을 유지합니다.
- `note`: 값이 있으면 홈페이지에서 학술지·학술대회 정보 바로 아래 같은 글꼴로 표시됩니다.
- `award_id`: 관련 수상이 있으면 `awards.json`의 ID, 없으면 빈 문자열. 수상 화면과 CV의 논문 정보는 이 연결을 통해 논문 데이터에서 가져옵니다.
- `doi`, `url`: 원문 링크.

`people.json`에서는 정확히 한 사람만 `"is_self": true`여야 합니다. 그 사람이 `author_ids`의 첫 번째이면 사이트가 자동으로 주저자 논문으로 판단하므로 별도의 저자 역할 필드는 필요하지 않습니다.

사람 객체의 `notes_en`, `notes_ko`는 언어별 문자열 배열입니다. 소속이나 간단한 정보를 한 줄에 하나씩 입력하면 논문 저자와 공저자 필터에서 선택한 언어의 툴팁으로 표시됩니다. 선택 언어의 배열이 비어 있으면 다른 언어 배열로 fallback하며, 정보가 없으면 두 배열을 모두 `[]`로 둡니다.

논문 주제의 명칭, 순서와 색상은 Profile-Editor의 **주제**에서 편집하며
`settings.json`의 `topics`에 저장됩니다. 각 주제의 선택 항목 `color`는
`"#0072b2"`처럼 `#RRGGBB` 형식으로 지정하며, 생략하면 기존 기본 색상을 사용합니다.
기타는 별도 주제가 아니라 빈 `topic`을 표시하는 fallback이며
`topic_fallback`에서 표시 명칭을 정합니다.

논문 탭 상단의 타임라인은 가로축에 출판 시점, 세로축에 네 가지 논문 유형을 표시합니다. 원의 색은 주제별로 구분하고, 주저자 논문은 채운 원, 공저자 논문은 테두리 원으로 표시합니다. 같은 날짜의 논문은 한 줄에 최대 세 개씩 배치하고 다음 줄은 왼쪽부터 정렬합니다. 날짜가 가까운 논문도 같은 3열 그리드의 빈칸을 채우도록 묶으며, 각 묶음 안에서는 주저자 논문을 먼저 배치하고 같은 저자 역할끼리는 날짜순으로 정렬합니다. 별도 묶음끼리 겹치면 아래 줄로 이동합니다. 날짜가 없는 논문은 ‘심사중’(영문: ‘under review’) 열에 표시하고, 연도·월까지만 입력한 날짜는 해당 기간의 시작점으로 계산합니다. 검색과 필터는 목록과 타임라인에 함께 반영되며, 원에 마우스를 올리거나 키보드로 초점을 맞추면 제목·저자와 정확한 날짜·게재지를 확인할 수 있습니다. 원을 선택하면 해당 논문으로 이동합니다.

논문에는 별도 ID를 입력하지 않습니다. BibTeX·CSL 내보내기에 규격상 필요한 인용 키는 DOI, URL 또는 서지정보를 바탕으로 빌드할 때 자동 생성됩니다. `awards.json`에는 논문 제목이나 저자 같은 서지정보를 중복 저장하지 않고 수상 자체 정보만 입력합니다.

## 프로젝트, 소프트웨어, 기록과 미디어

프로젝트, 소프트웨어, 기록은 이미지와 동영상을 함께 받을 수 있는 `media` 배열을 사용합니다.
프로젝트의 표시 순서는 JSON 배열에 적힌 순서이며 별도의 `order` 필드는 사용하지 않습니다. 소프트웨어는 공개 배포 단계(`release`) → 시험 공개 단계(`preview`) → 개발 단계(`development`)로 나누어 표시하고, 같은 단계 안에서는 JSON 배열 순서를 따릅니다. 프로젝트는 다른 데이터에서 참조되지 않으므로 `id`도 사용하지 않습니다.
프로젝트에는 `start_date`, `end_date`를 `YYYY-MM-DD` 형식으로 입력하고 `theme`에는 `settings.json`의 `topics`에 등록한 ID를 사용합니다. 날짜는 화면과 CV에서 `YYYY.MM.DD. – YYYY.MM.DD.` 형식으로 표시됩니다.
프로젝트와 소프트웨어 설명은 `summary` 문장 대신 `notes_en`, `notes_kr` 문자열 배열에 개조식으로 입력합니다. 각 배열의 항목은 홈페이지와 CV에서 하나의 bullet로 표시됩니다. 각 언어에 최소 한 항목을 입력하며, 항목 수와 순서는 언어별로 관리합니다. 예를 들어 국문 설명 세 항목을 영문에서는 한 항목으로 요약할 수 있습니다. 미디어 캡션은 `caption_en`, `caption_ko`를 한 쌍으로 입력하거나 둘 다 생략합니다.

```json
{
  "start_date": "2026-07-27",
  "end_date": "2027-07-26",
  "theme": "retrofit",
  "notes_en": [
    "Analysis of retrofit performance using dynamic simulation"
  ],
  "notes_kr": [
    "동적 시뮬레이션 기반 리모델링 성능 분석",
    "리모델링 시나리오별 성능 비교"
  ]
}
```

프로젝트 주제를 추가하거나 명칭·색상을 바꿀 때는 Profile-Editor의 **주제**에서
편집하거나 `settings.json`의 `topics`를 수정합니다. 선택 항목 `color`는
`#RRGGBB` 형식이며, 생략하면 기존 기본 색상을 사용합니다. 각 프로젝트의 `theme`은 이 목록에 등록된 ID여야
하며 `topic_fallback`은 예외 상황의 표시 명칭만 관리합니다.
프로젝트 탭의 상단 타임라인은 이 테마와 날짜를 이용해 자동 생성됩니다. 같은 테마에서 기간이 겹치는 과제는 여러 줄로 배치되며, 긴 전체 기간은 가로로 스크롤할 수 있습니다. 막대를 선택하면 해당 프로젝트 카드로 이동합니다.

```json
"media": [
  {
    "src": "projects/example.mp4",
    "poster": "projects/example-poster.jpg",
    "caption_en": "Simulation in progress",
    "caption_ko": "시뮬레이션 실행 화면"
  }
]
```

- 모든 `src`와 `poster`는 `data/media/` 기준 상대경로입니다.
- `.mp4`, `.webm`, `.ogv`, `.ogg`, `.mov`, `.m4v`는 동영상으로 자동 인식합니다.
- 확장자로 판단하기 어려우면 `"type": "video"` 또는 `"type": "image"`를 명시합니다.
- 프로젝트와 기록의 캡션은 선택 사항이며 소프트웨어 미디어는 영문·국문 캡션을 사용합니다.
- `board.json`의 `content_en`, `content_ko`는 문자열입니다.
- `board.json`의 `start_date`는 `YYYY-MM-DD` 형식이며 필수입니다. 여러 날에 걸친 항목은 `end_date`에 종료일을 입력하고, 하루짜리 항목은 빈 문자열로 둡니다. Board는 시작일 기준 최신순으로 표시됩니다.
- 소프트웨어의 `stage`는 `release`, `preview`, `development` 중 하나를 필수로 입력합니다.
- 소프트웨어의 `technologies`에는 언어, 프레임워크, 핵심 플랫폼 이름을 입력합니다. 지원되는 항목은 실제 로고로, 나머지는 텍스트 fallback으로 표시됩니다.

소프트웨어의 저장소, 배포 페이지, 보존 기록, 홈페이지는 `links` 배열 하나로 관리합니다. URL만 입력하면 도메인에 따라 플랫폼 이름과 아이콘을 자동으로 선택합니다.

```json
"links": [
  "https://github.com/example/example",
  "https://pypi.org/project/example/",
  "https://zenodo.org/records/1234567",
  "https://example.org"
]
```

자동 판별 대상은 다음과 같습니다.

- 코드 저장소: GitHub, GitLab, Bitbucket, Codeberg, SourceForge
- 패키지·앱 배포: food4Rhino, PyPI, npm, JSR, crates.io, NuGet, Maven Central, Anaconda, RubyGems, Packagist, pub.dev, Hex, CPAN, CRAN, Bioconductor, Go package, Flathub, Snap Store, Homebrew
- 컨테이너·모델 배포: Docker Hub, Quay, Hugging Face
- 연구 산출물·보존: Zenodo, Figshare, OSF, Software Heritage, DOI
- 위 목록에 없는 주소: 웹페이지

소프트웨어 링크와 연락처 아이콘은 `app/assets/icons/links/`에 모아 관리하며, GitHub는 같은 SVG 파일을 공유합니다. 기술 아이콘은 `app/assets/icons/technologies/`에 있습니다.

자동 이름 대신 별도 명칭이 필요한 일반 웹페이지는 객체로 적을 수 있습니다.

```json
"links": [
  {
    "url": "https://example.org/docs",
    "label_en": "Documentation",
    "label_ko": "문서"
  }
]
```

## 미리보기, 검증, 내보내기

Windows에서 다음 파일을 더블클릭할 수 있습니다.

| 도구 | 역할 |
|---|---|
| `app/tools/serve.bat` | 저장소 루트에서 로컬 서버를 열고 홈페이지 미리보기 실행 |
| `app/tools/validate.bat` | 데이터 구조, ID 참조, 미디어와 정적 파일 경로 검사 |
| `app/tools/build.bat` | 검증 후 서지 내보내기 파일과 보고서 재생성 |
| `app/tools/refresh-citations.bat` | OpenAlex 인용수 갱신과 선택 KCI 수동 수치 반영 |
| `app/tools/check-release.bat` | 프로그램·데이터 버전, 릴리스 노트와 업데이트 경계 검사 |

도구는 Go로 구현되어 있으며, 배치 파일은 `editor/setup.bat`으로 준비한 로컬 Go 또는
PATH의 Go를 사용합니다. 미리보기 서버는 Go HTTP 서버를 PowerShell에서 백그라운드로 실행합니다.
HTML 파일을 `file://`로 직접 열면 브라우저가 JSON 요청을 차단할 수 있으므로 `app/tools/serve.bat`을 사용하세요.
Go가 설치된 환경에서는 `editor/`에서 다음 명령을 직접 실행할 수도 있습니다.

```shell
go run ./cmd/site build
go run ./cmd/site validate
go run ./cmd/site serve --port 8000
go run ./cmd/citations
```

화면과 CV는 실행 시 `data/`를 직접 읽습니다. `app/tools/build.bat`이 생성하는 다음 파일은 화면 렌더링용 번들이 아니라 외부 활용을 위한 산출물입니다.

- `exports/publications.bib`
- `exports/publications.ris`
- `exports/publications.csl.json`
- `exports/build-report.json`

GitHub Actions는 push와 pull request마다 데이터 검증, 정적 경로 검사, JavaScript 문법 검사, 생성 파일의 최신 상태를 확인합니다.

## 권장 작업 흐름

```text
data 수정
→ app/tools/serve.bat으로 화면 확인
→ app/tools/validate.bat으로 검사
→ app/tools/build.bat으로 생성 파일 갱신
→ data/와 exports/ 변경 사항을 함께 commit·push
```

연락처와 경력뿐 아니라 이미지, 초록, 비공개 심사 상태 등 `data/`의 모든 내용은 배포 후 공개됩니다. push 전에 공개 가능한 정보인지 확인하세요.
