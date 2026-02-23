/**
 * 与 Go 后端 API 严格对齐的类型定义
 * 请求/响应字段名、URL 与 backend 一致
 */

// ========== Auth ==========
export interface RegisterRequest {
  username: string
  password: string
}

export interface RegisterResponse {
  message: string
  user_id: number
  username: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  message: string
  user_id: number
  username: string
  token: string
}

export interface ChangePassRequest {
  old_pass: string
  new_pass: string
}

// ========== Profile ==========
export interface UpdateProfileRequest {
  profile?: string
}

// ========== Posts ==========
/** GET /posts 查询参数 form: page, size, type, keyword */
export interface ListPostsQuery {
  page?: number
  size?: number
  type?: number  // 1=文章 2=问题
  keyword?: string
}

/** GET /posts 响应：list, total, page, page_size */
export interface ListPostsResponse {
  list: PostListItem[]
  total: number
  page: number
  page_size: number
}

/** 帖子列表项（无 json tag 时 Go 输出 PascalCase） */
export interface PostListItem {
  ID: number
  Type: number
  AuthorID: number
  AuthorName: string
  Title: string
  CreateAt: string
  UpdatedAt: string
}

export interface CreatePostRequest {
  type: number   // 1 | 2
  title: string
  content: string
  status?: number  // 0=发布 1=草稿
}

export interface CreatePostResponse {
  ok: boolean
  post: { ID: number; [key: string]: unknown }
}

export interface UpdatePostRequest {
  title?: string
  content?: string
  status?: number // 0=发布 1=草稿
}

/** GET /posts/:id 响应（PostDetailResp，仅 content 有 json tag） */
export interface PostDetailResponse {
  ID: number
  Type: number
  AuthorID: number
  AuthorName: string
  Title: string
  content: string
  Status?: number // 0=发布 1=草稿
  LikeCount: number
  CreatedAt: string
  UpdatedAt: string
}

// ========== Comments ==========
/** GET /posts/comments 查询参数 form: target_type, target_id, page, size */
export interface GetCommentsQuery {
  target_type: number  // 1=文章 2=问题
  target_id: number
  page?: number
  size?: number
}

/** GET /posts/comments 响应 data */
export interface GetCommentsResponse {
  message: string
  data: {
    comments: CommentItem[]
    total: number
    page: number
    size: number
  }
}

export interface CommentItem {
  id: number
  author_id: number
  author_name?: string
  content: string
  depth: number
  created_at: string
  like_count: number
  is_liked: boolean
}

/** POST /comments 请求 */
export interface PostCommentRequest {
  target_type: number  // 1=文章 2=问题 3=回复某条评论
  target_id: number
  content: string
}

/** GET /comments/:parent_id/replies 响应 data */
export interface GetRepliesResponse {
  message: string
  data: {
    replies: CommentItem[]
    total: number
  }
}

// ========== User ==========
/** GET /user/:id 响应 data (UserPublicInfo) */
export interface UserPublicInfo {
  id: number
  username: string
  avatar_url?: string
  profile?: string
  role: number
  is_vip: boolean
  vip_expires_at?: string
  posts: PostSummary[]
  post_total: number
  following_count: number
  followers_count: number
  page: number
  size: number
}

export interface PostSummary {
  id: number
  title: string
  created_at: string
  status: number
}

/** GET /users/followers/:id 或 /users/following/:id 查询 page, size */
export interface FollowListResponse {
  message: string
  data: {
    users: FollowUserInfo[]
    total: number
    page: number
    size: number
  }
}

export interface FollowUserInfo {
  id: number
  username: string
  avatar_url?: string
  profile?: string
  is_followed: boolean
}

// ========== Reactions & Favorites ==========
export interface LikeRequest {
  target_type: number  // 1=文章 2=问题 3=评论
  target_id: number
}

export interface ToggleReactionResponse {
  message: string
  status: boolean  // 操作后是否已点赞
}

export interface FavorRequest {
  target_type: number  // 1=文章 2=问题
  target_id: number
}

export interface ToggleFavoriteResponse {
  message: string
  data: { is_favorited: boolean }
}

// ========== Notifications ==========
export interface GetNotificationsResponse {
  message: string
  data: {
    notifications: NotificationItem[]
    total: number
    page: number
    size: number
  }
}

export interface GetUnreadCountResponse {
  count: number
}

export interface NotificationItem {
  id: number
  type: number
  actor_id: number
  actor_name?: string
  target_type?: number
  target_id?: number
  content: string
  is_read: boolean
  created_at: string
}

// ========== Favorites / Drafts ==========
export interface GetFavoritesResponse {
  message: string
  data: {
    favorites: FavoriteItem[]
    total: number
    page: number
    size: number
  }
}

export interface FavoriteItem {
  id: number
  type: number
  title: string
  created_at: string
}

export interface GetDraftResponse {
  message: string
  data: {
    drafts: PostListItem[]
    total: number
    page: number
    size: number
  }
}

// ========== Upload ==========
/** POST /avatar FormData field: avatar */
/** POST /upload/article-image FormData field: image */
export interface UploadArticleImageResponse {
  message: string
  image_url: string
}

export interface UploadAvatarResponse {
  avatar_url: string
}
