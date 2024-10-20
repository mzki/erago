package image

import (
	"bytes"
	"image"
	img "image"
	"image/png"
	"io"
	"os"
	"reflect"
	"testing"

	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

var (
	testImageData     img.Image = mustImage(loadPngImage("testdata/color.png"))
	testImageDataWebp img.Image = mustImage(loadWebpImage("testdata/color_resized.webp"))
	testImageLoader   Loader    = *NewLoader(4)
)

func mustImage(m image.Image, err error) img.Image {
	if err != nil {
		panic(err)
	}
	return m
}

func loadPngImage(file string) (img.Image, error) {
	fp, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	img, err := png.Decode(fp)
	return img, err
}

func dumpPngImage(file string, src img.Image) error {
	fp, err := os.Create(file)
	if err != nil {
		return err
	}
	defer fp.Close()
	if err := png.Encode(fp, src); err != nil {
		return err
	}
	return nil
}

func loadWebpImage(file string) (img.Image, error) {
	fp, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	img, err := webp.Decode(fp)
	return img, err
}

func TestNewLoader(t *testing.T) {
	type args struct {
		cachedSize int
	}
	tests := []struct {
		name    string
		args    args
		wantNil bool
	}{
		{"positive cache size", args{100}, false},
		{"negative cache size", args{-1}, false},
		{"zero cache size", args{0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLoader(tt.args.cachedSize); tt.wantNil && got != nil {
				t.Errorf("NewLoader() = %v, wantNil = %v", got, tt.wantNil)
			}
		})
	}
}

func TestLoader_Get(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name    string
		p       *Loader
		args    args
		want    img.Image
		wantErr bool
	}{
		{"Get normal", &testImageLoader, args{"testdata/color.png"}, testImageData, false},
		{"Get normal webp", &testImageLoader, args{"testdata/color_resized.webp"}, testImageDataWebp, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.Get(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("Loader.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Loader.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAutoLoadFile(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name    string
		args    args
		want    img.Image
		wantErr bool
	}{
		{"load test image", args{"testdata/color.png"}, testImageData, false},
		{"load test image webp", args{"testdata/color_resized.webp"}, testImageDataWebp, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AutoLoadFile(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("AutoLoadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AutoLoadFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAutoLoad(t *testing.T) {
	t.Skip("Not implemented")
	type args struct {
		r   io.Reader
		ext string
	}
	tests := []struct {
		name    string
		args    args
		want    img.Image
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AutoLoad(tt.args.r, tt.args.ext)
			if (err != nil) != tt.wantErr {
				t.Errorf("AutoLoad() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AutoLoad() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoader_GetWithOptions(t *testing.T) {
	wantImg, err := loadPngImage("testdata/color_resized.png")
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		file string
		opt  LoadOptions
	}
	tests := []struct {
		name    string
		p       *Loader
		args    args
		want    img.Image
		wantErr bool
	}{
		{"Get normal", &testImageLoader, args{"testdata/color.png", LoadOptions{img.Point{128, 64}}}, wantImg, false},
		{"zero resizedX", &testImageLoader, args{"testdata/color.png", LoadOptions{img.Point{0, 64}}}, wantImg, false},
		{"zero resizedY", &testImageLoader, args{"testdata/color.png", LoadOptions{img.Point{128, 0}}}, wantImg, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.p.GetWithOptions(tt.args.file, tt.args.opt)
			if (err != nil) != tt.wantErr {
				t.Errorf("Loader.GetWithOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Loader.GetWithOptions() = %v, want %v", got, tt.want)
			}

			// dump data
			//if err := dumpPngImage("testdata/color_resized.png", got); err != nil {
			//	t.Fatal(err)
			//}
		})
	}
}

func Test_createImageKey(t *testing.T) {
	type args struct {
		file string
		opt  LoadOptions
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal", args{"/path/to/image", LoadOptions{img.Point{11, 22}}}, "/path/to/image_11x22"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createImageKey(tt.args.file, tt.args.opt); got != tt.want {
				t.Errorf("createImageKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_resizeImage(t *testing.T) {
	t.Skip("Not implemented yet")
	type args struct {
		src img.Image
		opt LoadOptions
	}
	tests := []struct {
		name string
		args args
		want img.Image
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resizeImage(tt.args.src, tt.args.opt); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("resizeImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

var (
	benchReduceDstRect  = img.Rect(0, 0, 256, 128)
	benchEnlargeDstRect = img.Rect(0, 0, 1024, 512)
)

func BenchmarkResizeReduceApproxBilinear(b *testing.B) {
	testImage := testImageData
	dst := img.NewRGBA(benchReduceDstRect)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.ApproxBiLinear.Scale(dst, dst.Bounds(), testImage, testImage.Bounds(), draw.Over, nil)
	}
	b.StopTimer()
	dumpPngImage("testdata/bench_reduce_approx.png", dst)
}

func BenchmarkResizeReduceBilinear(b *testing.B) {
	testImage := testImageData
	dst := img.NewRGBA(benchReduceDstRect)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.BiLinear.Scale(dst, dst.Bounds(), testImage, testImage.Bounds(), draw.Over, nil)
	}
	b.StopTimer()
	dumpPngImage("testdata/bench_reduce_bilinear.png", dst)
}

func BenchmarkResizeReduceCatmullRom(b *testing.B) {
	testImage := testImageData
	dst := img.NewRGBA(benchReduceDstRect)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.CatmullRom.Scale(dst, dst.Bounds(), testImage, testImage.Bounds(), draw.Over, nil)
	}
	b.StopTimer()
	dumpPngImage("testdata/bench_reduce_catmullrom.png", dst)
}

func BenchmarkResizeEnlargeApproxBilinear(b *testing.B) {
	testImage := testImageData
	dst := img.NewRGBA(benchEnlargeDstRect)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.ApproxBiLinear.Scale(dst, dst.Bounds(), testImage, testImage.Bounds(), draw.Over, nil)
	}
	b.StopTimer()
	dumpPngImage("testdata/bench_enlarge_approx.png", dst)
}

func BenchmarkResizeEnlargeBilinear(b *testing.B) {
	testImage := testImageData
	dst := img.NewRGBA(benchEnlargeDstRect)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.BiLinear.Scale(dst, dst.Bounds(), testImage, testImage.Bounds(), draw.Over, nil)
	}
	b.StopTimer()
	dumpPngImage("testdata/bench_enlarge_biliear.png", dst)
}

func BenchmarkResizeEnlargeCatmullRom(b *testing.B) {
	testImage := testImageData
	dst := img.NewRGBA(benchEnlargeDstRect)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		draw.CatmullRom.Scale(dst, dst.Bounds(), testImage, testImage.Bounds(), draw.Over, nil)
	}
	b.StopTimer()
	dumpPngImage("testdata/bench_enlarge_catmullrom.png", dst)
}

func BenchmarkDecodePng(b *testing.B) {
	testData, _ := os.ReadFile("testdata/color_resized.png")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = png.Decode(bytes.NewBuffer(testData))
	}
	b.StopTimer()
}

func BenchmarkDecodeWebp(b *testing.B) {
	testData, _ := os.ReadFile("testdata/color_resized.webp")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = webp.Decode(bytes.NewBuffer(testData))
	}
	b.StopTimer()
}
