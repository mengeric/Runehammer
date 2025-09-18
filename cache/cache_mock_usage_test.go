package cache

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"
)

func TestMockCacheGeneratedHelpers(t *testing.T) {
	Convey("gomock generated cache helper coverage", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mock := NewMockCache(ctrl)

		mock.EXPECT().Set(gomock.Any(), "key", []byte("value"), time.Second).Return(nil).Times(1)
		mock.EXPECT().Get(gomock.Any(), "key").Return([]byte("value"), nil).Times(1)
		mock.EXPECT().Del(gomock.Any(), "key").Return(nil).Times(1)
		mock.EXPECT().Close().Return(nil).Times(1)

		ctx := context.Background()
		So(mock.Set(ctx, "key", []byte("value"), time.Second), ShouldBeNil)
		val, err := mock.Get(ctx, "key")
		So(err, ShouldBeNil)
		So(string(val), ShouldEqual, "value")
		So(mock.Del(ctx, "key"), ShouldBeNil)
		So(mock.Close(), ShouldBeNil)
	})
}
