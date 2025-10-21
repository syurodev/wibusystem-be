# Community Module - Placeholder

## Status: 🚧 NOT STARTED

The Community module will handle social features like comments, reviews, ratings, discussions, and user interactions.

## Planned Features

### Core Features
- **Comments**: User comments on novels/chapters
- **Reviews**: Detailed reviews with ratings
- **Ratings**: Star ratings for novels
- **Discussions**: Forum-like discussions
- **Notifications**: User activity notifications
- **Following**: Follow creators/novels
- **Bookmarks**: Save favorite novels/chapters

### Entities (Planned)

#### Comment
- ID, Content, UserID, NovelID/ChapterID
- ParentCommentID (for replies)
- Created/Updated timestamps
- Likes count
- Status (active, hidden, deleted)

#### Review
- ID, Title, Content, Rating
- UserID, NovelID
- Helpful votes
- Created/Updated timestamps
- Status

#### Rating
- ID, UserID, NovelID
- Score (1-5 stars)
- Created/Updated timestamps

#### Discussion
- ID, Title, Content
- UserID, Category
- Tags
- Views, Replies count
- Created/Updated timestamps

#### Notification
- ID, UserID, Type
- Content, Link
- Read status
- Created timestamp

#### Follow
- ID, FollowerID, FollowedID
- Type (user, novel, creator)
- Created timestamp

#### Bookmark
- ID, UserID, NovelID/ChapterID
- Created timestamp

## Target Structure

```
internal/modules/community/
├── domain/              # Domain entities
│   ├── comment.go
│   ├── review.go
│   ├── rating.go
│   ├── discussion.go
│   ├── notification.go
│   ├── follow.go
│   └── bookmark.go
├── repository/          # Repository interfaces
│   ├── comment_repository.go
│   ├── review_repository.go
│   ├── rating_repository.go
│   └── postgres/       # PostgreSQL implementations
├── service/             # Business logic
│   ├── comment_service.go
│   ├── review_service.go
│   ├── rating_service.go
│   ├── discussion_service.go
│   ├── notification_service.go
│   └── follow_service.go
├── handler/             # HTTP handlers
│   ├── http/
│   │   ├── comment_handler.go
│   │   ├── review_handler.go
│   │   ├── rating_handler.go
│   │   └── router.go
│   └── middleware/
└── dto/                 # Data transfer objects
    ├── comment.go
    ├── review.go
    └── rating.go
```

## API Endpoints (Estimated)

### Comments
- GET /api/v1/comments - List comments (filtered)
- GET /api/v1/comments/:id - Get comment
- POST /api/v1/comments - Create comment (auth)
- PUT /api/v1/comments/:id - Update comment (auth)
- DELETE /api/v1/comments/:id - Delete comment (auth)
- POST /api/v1/comments/:id/like - Like comment (auth)
- GET /api/v1/novels/:novelId/comments - Get novel comments
- GET /api/v1/chapters/:chapterId/comments - Get chapter comments

### Reviews
- GET /api/v1/reviews - List reviews
- GET /api/v1/reviews/:id - Get review
- POST /api/v1/reviews - Create review (auth)
- PUT /api/v1/reviews/:id - Update review (auth)
- DELETE /api/v1/reviews/:id - Delete review (auth)
- POST /api/v1/reviews/:id/helpful - Mark helpful (auth)
- GET /api/v1/novels/:novelId/reviews - Get novel reviews

### Ratings
- GET /api/v1/novels/:novelId/rating - Get average rating
- POST /api/v1/ratings - Create/update rating (auth)
- GET /api/v1/users/me/ratings - Get user ratings

### Discussions
- GET /api/v1/discussions - List discussions
- GET /api/v1/discussions/:id - Get discussion
- POST /api/v1/discussions - Create discussion (auth)
- PUT /api/v1/discussions/:id - Update discussion (auth)
- DELETE /api/v1/discussions/:id - Delete discussion (auth)
- POST /api/v1/discussions/:id/reply - Reply to discussion (auth)

### Notifications
- GET /api/v1/notifications - Get user notifications (auth)
- PUT /api/v1/notifications/:id/read - Mark as read (auth)
- DELETE /api/v1/notifications/:id - Delete notification (auth)
- PUT /api/v1/notifications/read-all - Mark all as read (auth)

### Following
- POST /api/v1/follow - Follow entity (auth)
- DELETE /api/v1/follow/:id - Unfollow (auth)
- GET /api/v1/users/me/following - Get following list (auth)
- GET /api/v1/users/:userId/followers - Get followers

### Bookmarks
- GET /api/v1/bookmarks - Get user bookmarks (auth)
- POST /api/v1/bookmarks - Create bookmark (auth)
- DELETE /api/v1/bookmarks/:id - Delete bookmark (auth)

**Estimated Total: 35-40 endpoints**

## Dependencies

Community module will depend on:
- **Identity module** for authentication/authorization
- **Catalog module** for novel/chapter references
- **Shared infrastructure** (database, config)

## Implementation Timeline

### Phase 1: Domain Layer (Days 1-2)
- [ ] Comment entity
- [ ] Review entity
- [ ] Rating entity
- [ ] Discussion entity
- [ ] Notification entity
- [ ] Follow entity
- [ ] Bookmark entity

### Phase 2: Repository Layer (Days 3-4)
- [ ] Repository interfaces
- [ ] PostgreSQL implementations
- [ ] Filtering and pagination
- [ ] Search functionality

### Phase 3: Service Layer (Days 5-7)
- [ ] Comment service
- [ ] Review service
- [ ] Rating service
- [ ] Discussion service
- [ ] Notification service
- [ ] Follow service
- [ ] Bookmark service

### Phase 4: Handler Layer (Days 8-9)
- [ ] HTTP handlers
- [ ] Route configuration
- [ ] Middleware integration
- [ ] Validation

### Phase 5: Testing (Day 10)
- [ ] Unit tests
- [ ] Integration tests
- [ ] API tests

**Estimated Total: 10-12 days**

## Features to Consider

### Moderation
- Report system for inappropriate content
- Admin moderation tools
- Auto-moderation (spam detection)

### Real-time
- WebSocket for live notifications
- Real-time comment updates
- Live discussion threads

### Gamification
- User reputation/karma
- Badges and achievements
- Leaderboards

### Analytics
- Popular content tracking
- User engagement metrics
- Trending discussions

## Database Design

### Tables Needed
- comments
- reviews
- ratings
- discussions
- discussion_replies
- notifications
- follows
- bookmarks
- comment_likes
- review_votes

## Priority

**Priority: LOW**

Community features are important for engagement but not critical for MVP. Should be implemented after:
1. ✅ Identity module (complete)
2. ⏳ Catalog module (next)
3. ⏳ Payment module (after catalog)
4. Then Community

## Notes

1. **Spam Prevention**: Implement rate limiting for comments/reviews
2. **Content Moderation**: Need tools to moderate user-generated content
3. **Notifications**: Consider using a queue system (Redis) for async processing
4. **Real-time**: May need WebSocket support for live features
5. **Caching**: Cache popular comments, trending discussions

---

**Last Updated:** January 2025  
**Status:** Planning phase  
**Priority:** LOW  
**Next Action:** Wait for Catalog and Payment modules completion