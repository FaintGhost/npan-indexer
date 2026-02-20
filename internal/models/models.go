package models

type ItemType string

const (
	ItemTypeFile   ItemType = "file"
	ItemTypeFolder ItemType = "folder"
)

type NpanFolder struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	ParentID   int64  `json:"parent_id"`
	ModifiedAt int64  `json:"modified_at,omitempty"`
	InTrash    bool   `json:"in_trash,omitempty"`
	IsDeleted  bool   `json:"is_deleted,omitempty"`
}

type NpanFile struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	ParentID   int64  `json:"parent_id"`
	Size       int64  `json:"size,omitempty"`
	ModifiedAt int64  `json:"modified_at,omitempty"`
	CreatedAt  int64  `json:"created_at,omitempty"`
	SHA1       string `json:"sha1,omitempty"`
	InTrash    bool   `json:"in_trash,omitempty"`
	IsDeleted  bool   `json:"is_deleted,omitempty"`
}

type FolderChildrenPage struct {
	Folders      []NpanFolder `json:"folders"`
	Files        []NpanFile   `json:"files"`
	PageID       int64        `json:"page_id"`
	PageCount    int64        `json:"page_count"`
	PageCapacity int64        `json:"page_capacity"`
	TotalCount   int64        `json:"total_count,omitempty"`
}

type IndexDocument struct {
	DocID      string   `json:"doc_id"`
	SourceID   int64    `json:"source_id"`
	Type       ItemType `json:"type"`
	Name       string   `json:"name"`
	PathText   string   `json:"path_text"`
	ParentID   int64    `json:"parent_id"`
	ModifiedAt int64    `json:"modified_at"`
	CreatedAt  int64    `json:"created_at"`
	Size       int64    `json:"size"`
	SHA1       string   `json:"sha1"`
	InTrash    bool     `json:"in_trash"`
	IsDeleted  bool     `json:"is_deleted"`
}

type CrawlStats struct {
	FoldersVisited int64 `json:"foldersVisited"`
	FilesIndexed   int64 `json:"filesIndexed"`
	PagesFetched   int64 `json:"pagesFetched"`
	FailedRequests int64 `json:"failedRequests"`
	StartedAt      int64 `json:"startedAt"`
	EndedAt        int64 `json:"endedAt"`
}

type RetryPolicyOptions struct {
	MaxRetries  int `json:"maxRetries"`
	BaseDelayMS int `json:"baseDelayMs"`
	MaxDelayMS  int `json:"maxDelayMs"`
	JitterMS    int `json:"jitterMs"`
}

type DownloadURLResult struct {
	DownloadURL string `json:"download_url"`
	ExpiresAt   int64  `json:"expires_at,omitempty"`
}

type NpanDepartment struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type RootSyncProgress struct {
	RootFolderID     int64      `json:"rootFolderId"`
	CheckpointFile   string     `json:"checkpointFile"`
	Status           string     `json:"status"`
	Stats            CrawlStats `json:"stats"`
	CurrentFolderID  *int64     `json:"currentFolderId,omitempty"`
	CurrentPageID    *int64     `json:"currentPageId,omitempty"`
	CurrentPageCount *int64     `json:"currentPageCount,omitempty"`
	QueueLength      *int64     `json:"queueLength,omitempty"`
	UpdatedAt        int64      `json:"updatedAt"`
	Error            string     `json:"error,omitempty"`
}

type SyncProgressState struct {
	Status             string                       `json:"status"`
	StartedAt          int64                        `json:"startedAt"`
	UpdatedAt          int64                        `json:"updatedAt"`
	MeiliHost          string                       `json:"meiliHost"`
	MeiliIndex         string                       `json:"meiliIndex"`
	CheckpointTemplate string                       `json:"checkpointTemplate"`
	Roots              []int64                      `json:"roots"`
	CompletedRoots     []int64                      `json:"completedRoots"`
	ActiveRoot         *int64                       `json:"activeRoot,omitempty"`
	AggregateStats     CrawlStats                   `json:"aggregateStats"`
	RootProgress       map[string]*RootSyncProgress `json:"rootProgress"`
	LastError          string                       `json:"lastError,omitempty"`
}

type CrawlCheckpoint struct {
	Queue           []int64 `json:"queue"`
	CurrentFolderID *int64  `json:"currentFolderId,omitempty"`
	CurrentPageID   *int64  `json:"currentPageId,omitempty"`
}

type SyncState struct {
	LastSyncTime int64 `json:"lastSyncTime"`
}

type LocalSearchParams struct {
	Query          string
	Type           string
	Page           int64
	PageSize       int64
	ParentID       *int64
	UpdatedAfter   *int64
	UpdatedBefore  *int64
	IncludeDeleted bool
}

type RemoteSearchParams struct {
	QueryWords       string
	Type             string
	PageID           int64
	QueryFilter      string
	SearchInFolder   *int64
	UpdatedTimeRange string
}

type RemoteSearchItem struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type RemoteSearchResponse struct {
	Files        []RemoteSearchItem `json:"files"`
	Folders      []RemoteSearchItem `json:"folders"`
	TotalCount   int64              `json:"total_count"`
	PageID       int64              `json:"page_id"`
	PageCapacity int64              `json:"page_capacity"`
	PageCount    int64              `json:"page_count"`
}
