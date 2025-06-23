# Flakeguard Architecture Diagrams

## 1. High-Level System Architecture

```mermaid
graph TB
    subgraph "User Interface Layer"
        CLI[Flakeguard CLI]
        CI[CI/CD Systems]
        Dev[Developers]
    end
    
    subgraph "Core Application Layer"
        Detect[Detect Mode]
        Guard[Guard Mode]
        Report[Report Engine]
        Analysis[Flakiness Analysis]
    end
    
    subgraph "External Tool Integration"
        GoTestSum[gotestsum]
        GoTest[go test]
    end
    
    subgraph "External Services"
        GitHub[GitHub API]
        Splunk[Splunk HEC]
        Jira[Jira API]
        DX[DX Webhook]
        Slack[Slack Webhook]
    end
    
    subgraph "Data Storage"
        JSON[JSON Files]
        Logs[Log Files]
        Cache[Package Cache]
    end
    
    CLI --> Detect
    CLI --> Guard
    CI --> CLI
    Dev --> CLI
    
    Detect --> GoTestSum
    Guard --> GoTestSum
    GoTestSum --> GoTest
    
    Detect --> Report
    Guard --> Report
    Report --> Analysis
    
    Report --> GitHub
    Report --> Splunk
    Report --> Jira
    Report --> DX
    Report --> Slack
    
    GoTestSum --> JSON
    Report --> Logs
    Analysis --> Cache
```

## 2. Detailed Component Dataflow

```mermaid
flowchart TD
    subgraph "Input Processing"
        A[Command Line Args] --> B[Argument Parser]
        B --> C[Flag Validation]
        C --> D[Configuration Setup]
    end
    
    subgraph "Test Execution Engine"
        D --> E[Test Runner]
        E --> F[gotestsum Integration]
        F --> G[go test Execution]
        G --> H[JSON Output Generation]
    end
    
    subgraph "Result Processing"
        H --> I[JSON Parser]
        I --> J[Test Result Aggregation]
        J --> K[Flakiness Analysis]
        K --> L[Report Generation]
    end
    
    subgraph "Output & Integration"
        L --> M[Console Output]
        L --> N[File Output]
        L --> O[External Service Integration]
        O --> P[GitHub API]
        O --> Q[Splunk]
        O --> R[Jira]
        O --> S[Slack]
    end
    
    subgraph "Supporting Systems"
        T[Logging System]
        U[Error Handling]
        V[Exit Code Management]
        W[Git Repository Info]
    end
    
    D --> T
    E --> U
    L --> V
    D --> W
```

## 3. Detect Mode Detailed Flow

```mermaid
flowchart TD
    Start([Start Detect]) --> Parse[Parse Arguments]
    Parse --> Validate[Validate Flags]
    Validate --> Setup[Setup Logging & Config]
    Setup --> GetRepo[Get Repository Info]
    GetRepo --> Init[Initialize Test Run]
    
    Init --> Loop{For Each Run}
    Loop --> RunTest[Run gotestsum]
    RunTest --> StoreJSON[Store JSON Results]
    StoreJSON --> CheckDuration{Check Duration Target}
    CheckDuration -->|Continue| Loop
    CheckDuration -->|Stop| ProcessResults
    
    ProcessResults --> LoadJSON[Load All JSON Files]
    LoadJSON --> ParseResults[Parse Test Results]
    ParseResults --> AnalyzeFlakes[Analyze Flakiness]
    AnalyzeFlakes --> GenerateReport[Generate Report]
    
    GenerateReport --> Console[Output to Console]
    GenerateReport --> Files[Write Report Files]
    GenerateReport --> External{Send to External Services?}
    
    External -->|Yes| Splunk[Send to Splunk]
    External -->|Yes| Jira[Send to Jira]
    External -->|Yes| GitHub[Update GitHub]
    External -->|No| End
    
    Splunk --> End
    Jira --> End
    GitHub --> End
    Console --> End
    Files --> End
    
    End([End Detect])
```

## 4. Guard Mode Flow (Planned)

```mermaid
flowchart TD
    Start([Start Guard]) --> Setup[Setup & Validation]
    Setup --> GetChanges[Get PR Changes]
    GetChanges --> IdentifyTests[Identify New/Modified Tests]
    IdentifyTests --> RunInitial[Run Initial Test Suite]
    
    RunInitial --> CheckFailures{Any Failures?}
    CheckFailures -->|Yes| RetryTests[Retry Failing Tests]
    CheckFailures -->|No| CheckNewTests
    
    RetryTests --> RetryCount{Retry Count < Max?}
    RetryCount -->|Yes| RetryTests
    RetryCount -->|No| CheckNewTests
    
    CheckNewTests --> NewTests{New Tests Found?}
    NewTests -->|Yes| DetectNew[Run Detect on New Tests]
    NewTests -->|No| AllowMerge
    
    DetectNew --> NewFlakes{New Flakes Detected?}
    NewFlakes -->|Yes| BlockPR[Block PR]
    NewFlakes -->|No| AllowMerge
    
    BlockPR --> CreateTicket[Create Jira Ticket]
    CreateTicket --> CommentPR[Comment on PR]
    CommentPR --> End
    
    AllowMerge --> End([Allow Merge])
```

## 5. Data Structures and Relationships

```mermaid
erDiagram
    TestResult {
        string name
        string package
        string path
        time time_run
        bool package_panic
        bool panic
        bool timeout
        bool race
        bool skipped
        float64 pass_ratio
        int runs
        int failures
        int successes
        int skips
        array failing_run_numbers
        array durations
        map outputs
    }
    
    TestRunInfo {
        string repo_url
        string repo_owner
        string repo_name
        string default_branch
        bool on_default_branch
        string head_branch
        string head_commit
        string base_branch
        string base_commit
        string github_event
        string github_workflow
        string github_run_id
        string github_run_number
    }
    
    PackageInfo {
        string import_path
        string name
        string dir
        array go_files
        array test_go_files
        array xtest_go_files
        string module
        bool is_command
    }
    
    ActionsEnv {
        string action
        string actor
        string repository
        string ref
        string sha
        string workflow
        int64 run_id
        int run_number
        string token
    }
    
    TestResult ||--|| TestRunInfo : "belongs to"
    PackageInfo ||--o{ TestResult : "contains"
    ActionsEnv ||--o| TestRunInfo : "provides context"
```

## 6. External Service Integration Flow

```mermaid
sequenceDiagram
    participant FG as Flakeguard
    participant Config as Configuration
    participant Splunk as Splunk HEC
    participant Jira as Jira API
    participant GitHub as GitHub API
    participant Slack as Slack Webhook
    
    FG->>Config: Load service configurations
    Config-->>FG: Service endpoints & tokens
    
    FG->>FG: Generate test report
    
    alt Splunk configured
        FG->>Splunk: POST /services/collector
        Splunk-->>FG: 200 OK
    end
    
    alt Jira configured
        FG->>Jira: POST /rest/api/2/issue
        Jira-->>FG: Issue created
    end
    
    alt GitHub Actions environment
        FG->>GitHub: Update workflow status
        GitHub-->>FG: Status updated
    end
    
    alt Slack configured
        FG->>Slack: POST webhook URL
        Slack-->>FG: Message sent
    end
    
    FG->>FG: Log integration results
```

## 7. Error Handling and Recovery Flow

```mermaid
flowchart TD
    Start([Operation Start]) --> Try[Try Operation]
    Try --> Success{Success?}
    Success -->|Yes| Continue[Continue Processing]
    Success -->|No| ErrorType{Error Type?}
    
    ErrorType -->|Network| RetryNetwork[Retry with Backoff]
    ErrorType -->|Validation| LogValidation[Log Validation Error]
    ErrorType -->|System| LogSystem[Log System Error]
    ErrorType -->|Unknown| LogUnknown[Log Unknown Error]
    
    RetryNetwork --> RetryCount{Retry Count < Max?}
    RetryCount -->|Yes| Try
    RetryCount -->|No| FatalNetwork[Fatal Network Error]
    
    LogValidation --> ExitValidation[Exit with Validation Code]
    LogSystem --> ExitSystem[Exit with System Code]
    LogUnknown --> ExitUnknown[Exit with Unknown Code]
    FatalNetwork --> ExitNetwork[Exit with Network Code]
    
    Continue --> End([Operation Complete])
    ExitValidation --> End
    ExitSystem --> End
    ExitUnknown --> End
    ExitNetwork --> End
```

## 8. Performance and Resource Management

```mermaid
flowchart TD
    subgraph "Resource Management"
        A[Memory Pool] --> B[Goroutine Pool]
        B --> C[File Handle Pool]
        C --> D[Network Connection Pool]
    end
    
    subgraph "Performance Optimization"
        E[Package Caching] --> F[Test Result Caching]
        F --> G[Parallel Processing]
        G --> H[Streaming JSON Parsing]
    end
    
    subgraph "Monitoring"
        I[Memory Usage] --> J[CPU Usage]
        J --> K[Network I/O]
        K --> L[Disk I/O]
    end
    
    subgraph "Throttling"
        M[Rate Limiting] --> N[Concurrency Control]
        N --> O[Resource Limits]
        O --> P[Timeout Management]
    end
    
    A --> E
    B --> G
    C --> H
    D --> M
    
    E --> I
    G --> J
    H --> K
    M --> L
```

## 9. Security and Authentication Flow

```mermaid
flowchart TD
    Start([Authentication Start]) --> CheckEnv{Check Environment}
    CheckEnv -->|GitHub Actions| UseGithubToken[Use GITHUB_TOKEN]
    CheckEnv -->|Local| CheckConfig[Check Config Files]
    CheckEnv -->|CLI Args| UseCLIToken[Use CLI Token]
    
    UseGithubToken --> ValidateToken[Validate Token]
    CheckConfig --> LoadConfig[Load Config File]
    LoadConfig --> ValidateToken
    UseCLIToken --> ValidateToken
    
    ValidateToken --> Valid{Token Valid?}
    Valid -->|Yes| SetClient[Set API Client]
    Valid -->|No| ErrorAuth[Authentication Error]
    
    SetClient --> TestConnection[Test API Connection]
    TestConnection --> Connected{Connected?}
    Connected -->|Yes| Ready[Ready for Operations]
    Connected -->|No| ErrorConnection[Connection Error]
    
    ErrorAuth --> LogError[Log Authentication Error]
    ErrorConnection --> LogError
    LogError --> Exit[Exit with Error Code]
    
    Ready --> End([Authentication Complete])
    Exit --> End
```

---

*These diagrams provide detailed visual representations of Flakeguard's architecture, dataflow, and system interactions. They complement the comprehensive documentation and help understand the system's complexity and design decisions.* 