package memory

import (
	"errors"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

func Test_filePersist_addData(t *testing.T) {
	tests := []struct {
		name    string
		mockSet func(writerFactory *MockFactoryWriterInt, writer *MockWriterInt)
		wantErr bool
	}{
		{
			name: "new writer error",
			mockSet: func(writerFactory *MockFactoryWriterInt, writer *MockWriterInt) {
				writerFactory.EXPECT().
					NewWriter(mock.AnythingOfType("string")).
					Return(nil, errors.New("factory error")).
					Once()
			},
			wantErr: true,
		},
		{
			name: "write error",
			mockSet: func(writerFactory *MockFactoryWriterInt, writer *MockWriterInt) {
				writerFactory.EXPECT().
					NewWriter(mock.AnythingOfType("string")).
					Return(writer, nil).
					Once()

				writer.EXPECT().Close().Return(nil).Once()
				writer.EXPECT().
					Write(mock.AnythingOfType("[]uint8")).
					Return(0, errors.New("write error")).
					Once()
			},
			wantErr: true,
		},
		{
			name: "write success",
			mockSet: func(writerFactory *MockFactoryWriterInt, writer *MockWriterInt) {
				writerFactory.EXPECT().
					NewWriter(mock.AnythingOfType("string")).
					Return(writer, nil).
					Once()

				writer.EXPECT().Close().Return(nil).Once()
				writer.EXPECT().
					Write(mock.AnythingOfType("[]uint8")).
					Return(0, nil).
					Once()
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writerFactory := NewMockFactoryWriterInt(t)
			writer := NewMockWriterInt(t)
			tt.mockSet(writerFactory, writer)
			f := &filePersist{
				filePath:      "",
				writerFactory: writerFactory,
				readerFactory: nil,
			}
			if err := f.addData("key", "value", time.Now(), 1, true); (err != nil) != tt.wantErr {
				t.Errorf("addData() error = %v, wantErr %v", err, tt.wantErr)
			}
			writerFactory.AssertExpectations(t)
			writer.AssertExpectations(t)
		})
	}
}

var successString = "{\"short_url\":\"9W0wsMxE\",\"original_url\":\"original URL\", \"created_at\":\"2026-02-14T00:39:10+03:00\", \"user_id\": 1, \"active\": true}\n{\"short\":\"10BfwUi\",\"original\":\"original_url\"}\n{\"short_url\":\"Pc7V6XI3\",\"original_url\":\"original URL\", \"created_at\":\"2026-02-14T00:39:10+03:00\", \"user_id\": 2, \"active\": true}"

func Test_filePersist_getData(t *testing.T) {
	tests := []struct {
		name    string
		want    []urlData
		mockSet func(writerFactory *MockFactoryReaderInt, writer *MockReaderInt)
		wantErr bool
	}{
		{
			name: "new reader error",
			want: nil,
			mockSet: func(writerFactory *MockFactoryReaderInt, writer *MockReaderInt) {
				writerFactory.EXPECT().
					NewReader(mock.AnythingOfType("string")).
					Return(nil, errors.New("factory error")).
					Once()
			},
			wantErr: true,
		},
		{
			name: "consumer nil",
			want: nil,
			mockSet: func(writerFactory *MockFactoryReaderInt, writer *MockReaderInt) {
				writerFactory.EXPECT().
					NewReader(mock.AnythingOfType("string")).
					Return(nil, nil).
					Once()
			},
			wantErr: false,
		},
		{
			name: "read err",
			want: nil,
			mockSet: func(writerFactory *MockFactoryReaderInt, writer *MockReaderInt) {
				writerFactory.EXPECT().
					NewReader(mock.AnythingOfType("string")).
					Return(writer, nil).
					Once()

				writer.EXPECT().Close().Return(nil).Once()
				writer.EXPECT().
					Read(mock.AnythingOfType("[]uint8")).
					Return(0, errors.New("read error")).
					Once()
			},
			wantErr: true,
		},
		{
			name: "read empty",
			want: nil,
			mockSet: func(writerFactory *MockFactoryReaderInt, writer *MockReaderInt) {
				writerFactory.EXPECT().
					NewReader(mock.AnythingOfType("string")).
					Return(writer, nil).
					Once()

				writer.EXPECT().Close().Return(nil).Once()
				writer.EXPECT().
					Read(mock.AnythingOfType("[]uint8")).
					Return(0, io.EOF).
					Run(func(p []byte) {
						copy(p, []byte(""))
					}).
					Once()
			},
			wantErr: false,
		},
		{
			name: "read new line",
			want: nil,
			mockSet: func(writerFactory *MockFactoryReaderInt, writer *MockReaderInt) {
				writerFactory.EXPECT().
					NewReader(mock.AnythingOfType("string")).
					Return(writer, nil).
					Once()

				writer.EXPECT().Close().Return(nil).Once()
				writer.EXPECT().
					Read(mock.AnythingOfType("[]uint8")).
					Return(len("\n"), io.EOF).
					Run(func(p []byte) {
						copy(p, []byte("\n"))
					}).
					Once()
			},
			wantErr: false,
		},
		{
			name: "read not json line",
			want: nil,
			mockSet: func(writerFactory *MockFactoryReaderInt, writer *MockReaderInt) {
				writerFactory.EXPECT().
					NewReader(mock.AnythingOfType("string")).
					Return(writer, nil).
					Once()

				writer.EXPECT().Close().Return(nil).Once()
				writer.EXPECT().
					Read(mock.AnythingOfType("[]uint8")).
					Return(len("qwerty\n"), io.EOF).
					Run(func(p []byte) {
						copy(p, []byte("qwerty\n"))
					}).
					Once()
			},
			wantErr: false,
		},
		{
			name: "incorrect json",
			want: nil,
			mockSet: func(writerFactory *MockFactoryReaderInt, writer *MockReaderInt) {
				writerFactory.EXPECT().
					NewReader(mock.AnythingOfType("string")).
					Return(writer, nil).
					Once()

				writer.EXPECT().Close().Return(nil).Once()
				writer.EXPECT().
					Read(mock.AnythingOfType("[]uint8")).
					Return(len("\"{\"short\":\"9W0wsMxE\",\"original\":\"incorrect json\"}\n"), io.EOF).
					Run(func(p []byte) {
						copy(p, []byte("\"{\"short\":\"9W0wsMxE\",\"original\":\"incorrect json\"}\n"))
					}).
					Once()
			},
			wantErr: false,
		},

		{
			name: "read success",
			want: []urlData{
				{"9W0wsMxE", "original URL", "2026-02-14T00:39:10+03:00", 1, true},
				{"Pc7V6XI3", "original URL", "2026-02-14T00:39:10+03:00", 2, true},
			},
			mockSet: func(writerFactory *MockFactoryReaderInt, writer *MockReaderInt) {
				writerFactory.EXPECT().
					NewReader(mock.AnythingOfType("string")).
					Return(writer, nil).
					Once()

				writer.EXPECT().Close().Return(nil).Once()
				writer.EXPECT().
					Read(mock.AnythingOfType("[]uint8")).
					Return(len(successString), io.EOF).
					Run(func(p []byte) {
						copy(p, []byte(successString))
					}).
					Once()
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readerFactory := NewMockFactoryReaderInt(t)
			reader := NewMockReaderInt(t)
			tt.mockSet(readerFactory, reader)

			f := &filePersist{
				filePath:      "",
				writerFactory: nil,
				readerFactory: readerFactory,
			}

			got, err := f.getData()
			if (err != nil) != tt.wantErr {
				t.Errorf("getData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getData() got = %v, want %v", got, tt.want)
			}

			readerFactory.AssertExpectations(t)
			reader.AssertExpectations(t)
		})
	}
}
