package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "strings"
    "os"
    "gopkg.in/h2non/bimg.v1"
)

// OperationsMap defines the allowed image transformation operations listed by name.
// Used for pipeline image processing.
var OperationsMap = map[string]Operation{
    "crop":      Crop,
    "resize":    Resize,
    "enlarge":   Enlarge,
    "extract":   Extract,
    "rotate":    Rotate,
    "flip":      Flip,
    "flop":      Flop,
    "thumbnail": Thumbnail,
    "zoom":      Zoom,
    "convert":   Convert,
    "watermark": Watermark,
    "watermarkImage": watermarkImage,
    "blur":      GaussianBlur,
    "smartcrop": SmartCrop,
    "fit":       Fit,
}

// Image stores an image binary buffer and its MIME type
type Image struct {
    Body []byte
    Mime string
}

// Operation implements an image transformation runnable interface
type Operation func([]byte, ImageOptions) (Image, error)

// Run performs the image transformation
func (o Operation) Run(buf []byte, opts ImageOptions) (Image, error) {
    fmt.Printf("7. size = %+v \n", len(buf))
    return o(buf, opts)
}

// ImageInfo represents an image details and additional metadata
type ImageInfo struct {
    Width       int    `json:"width"`
    Height      int    `json:"height"`
    Type        string `json:"type"`
    Space       string `json:"space"`
    Alpha       bool   `json:"hasAlpha"`
    Profile     bool   `json:"hasProfile"`
    Channels    int    `json:"channels"`
    Orientation int    `json:"orientation"`
}

func Info(buf []byte, o ImageOptions) (Image, error) {
    // We're not handling an image here, but we reused the struct.
    // An interface will be definitively better here.
    image := Image{Mime: "application/json"}

    meta, err := bimg.Metadata(buf)
    if err != nil {
        return image, NewError("Cannot retrieve image metadata: %s"+err.Error(), BadRequest)
    }

    info := ImageInfo{
        Width:       meta.Size.Width,
        Height:      meta.Size.Height,
        Type:        meta.Type,
        Space:       meta.Space,
        Alpha:       meta.Alpha,
        Profile:     meta.Profile,
        Channels:    meta.Channels,
        Orientation: meta.Orientation,
    }

    body, _ := json.Marshal(info)
    image.Body = body

    return image, nil
}

func Resize(buf []byte, o ImageOptions) (Image, error) {
    if o.Width == 0 && o.Height == 0 {
        return Image{}, NewError("Missing required param: height or width", BadRequest)
    }

    opts := BimgOptions(o)
    opts.Embed = true

    if o.NoCrop == false {
        opts.Crop = true
    }

    //return Process(buf, opts)
    return AddWatermarkImage(o, buf, opts)
}

func Fit(buf []byte, o ImageOptions) (Image, error) {

    if o.Width == 0 || o.Height == 0 {
        return Image{}, NewError("Missing required params: height, width", BadRequest)
    }

    dims, err := bimg.Size(buf)
    if err != nil {
        return Image{}, err
    }

    // if original < fit
    if dims.Width < o.Width || dims.Height < o.Height {
        o.Width = dims.Width
        o.Height = dims.Height
    } else {

        // if input ratio > output ratio
        // (calculation multiplied through by denominators to avoid float division)
        if dims.Width*o.Height > o.Width*dims.Height {
            // constrained by width
            if dims.Width != 0 {
                o.Height = o.Width * dims.Height / dims.Width
            }
        } else {
            // constrained by height
            if dims.Height != 0 {
                o.Width = o.Height * dims.Width / dims.Height
            }
        }
    }


    fmt.Printf("2. FIT quality = %d\n", o.Quality)
    opts := BimgOptions(o)
    opts.Embed = true


    return AddWatermarkImage(o, buf, opts)
}

func AddWatermarkImage (o ImageOptions, buf2 []byte, opts bimg.Options)(Image, error){

//    fmt.Printf("2. %+v\n",opts);

    if o.CustomWatermark != "" {

        fmt.Printf("3. pre opts %+v  \n", opts)

/*        swapimage, swaperr :=  Process(buf2, opts)
        if swaperr != nil {
            return Image{}, swaperr
        }
*/
        o.Image = o.CustomWatermark;

        if o.WatermarkOpacity != 0 {
            o.Opacity = o.WatermarkOpacity
        } else {
            o.Opacity = 1.2
        }

        fmt.Printf("3. Add watermark quality = %d and size = %d \n", o.Quality,len(buf2))
  //      fmt.Printf("3.5 Add watermark quality = %d and size = %d \n", o.Quality,len(swapimage.Body))

        //return watermarkImage(swapimage.Body, o)
        return watermarkImage(buf2, o)
    }else{
        fmt.Printf("3. bis Add watermark quality = %d and size = %d \n", o.Quality,len(buf2))
        return Process(buf2, opts)
    }

}

func Enlarge(buf []byte, o ImageOptions) (Image, error) {
    if o.Width == 0 || o.Height == 0 {
        return Image{}, NewError("Missing required params: height, width", BadRequest)
    }

    opts := BimgOptions(o)
    opts.Enlarge = true

    if o.NoCrop == false {
        opts.Crop = true
    }

    return Process(buf, opts)
}

func Extract(buf []byte, o ImageOptions) (Image, error) {
    if o.AreaWidth == 0 || o.AreaHeight == 0 {
        return Image{}, NewError("Missing required params: areawidth or areaheight", BadRequest)
    }

    opts := BimgOptions(o)
    opts.Top = o.Top
    opts.Left = o.Left
    opts.AreaWidth = o.AreaWidth
    opts.AreaHeight = o.AreaHeight

    return Process(buf, opts)
}

func Crop(buf []byte, o ImageOptions) (Image, error) {
    if o.Width == 0 && o.Height == 0 {
        return Image{}, NewError("Missing required param: height or width", BadRequest)
    }

    opts := BimgOptions(o)
    opts.Crop = true
    //return Process(buf, opts)
    return AddWatermarkImage(o, buf, opts)
}

func SmartCrop(buf []byte, o ImageOptions) (Image, error) {
    if o.Width == 0 && o.Height == 0 {
        return Image{}, NewError("Missing required param: height or width", BadRequest)
    }

    opts := BimgOptions(o)
    opts.Crop = true
    opts.Gravity = bimg.GravitySmart
    //return Process(buf, opts)
    return AddWatermarkImage(o, buf, opts)
}

func Rotate(buf []byte, o ImageOptions) (Image, error) {
    if o.Rotate == 0 {
        return Image{}, NewError("Missing required param: rotate", BadRequest)
    }

    opts := BimgOptions(o)
    return Process(buf, opts)
}

func Flip(buf []byte, o ImageOptions) (Image, error) {
    opts := BimgOptions(o)
    opts.Flip = true
    return Process(buf, opts)
}

func Flop(buf []byte, o ImageOptions) (Image, error) {
    opts := BimgOptions(o)
    opts.Flop = true
    return Process(buf, opts)
}

func Thumbnail(buf []byte, o ImageOptions) (Image, error) {
    if o.Width == 0 && o.Height == 0 {
        return Image{}, NewError("Missing required params: width or height", BadRequest)
    }

    return Process(buf, BimgOptions(o))
}

func Zoom(buf []byte, o ImageOptions) (Image, error) {
    if o.Factor == 0 {
        return Image{}, NewError("Missing required param: factor", BadRequest)
    }

    opts := BimgOptions(o)

    if o.Top > 0 || o.Left > 0 {
        if o.AreaWidth == 0 && o.AreaHeight == 0 {
            return Image{}, NewError("Missing required params: areawidth, areaheight", BadRequest)
        }

        opts.Top = o.Top
        opts.Left = o.Left
        opts.AreaWidth = o.AreaWidth
        opts.AreaHeight = o.AreaHeight

        if o.NoCrop == false {
            opts.Crop = true
        }
    }

    opts.Zoom = o.Factor
    return Process(buf, opts)
}

func Convert(buf []byte, o ImageOptions) (Image, error) {
    if o.Type == "" {
        return Image{}, NewError("Missing required param: type", BadRequest)
    }
    if ImageType(o.Type) == bimg.UNKNOWN {
        return Image{}, NewError("Invalid image type: "+o.Type, BadRequest)
    }
    opts := BimgOptions(o)

    return Process(buf, opts)
}

func Watermark(buf []byte, o ImageOptions) (Image, error) {
    if o.Text == "" {
        return Image{}, NewError("Missing required param: text", BadRequest)
    }

    opts := BimgOptions(o)
    opts.Watermark.DPI = o.DPI
    opts.Watermark.Text = o.Text
    opts.Watermark.Font = o.Font
    opts.Watermark.Margin = o.Margin
    opts.Watermark.Width = o.TextWidth
    opts.Watermark.Opacity = o.Opacity
    opts.Watermark.NoReplicate = o.NoReplicate

    if len(o.Color) > 2 {
        opts.Watermark.Background = bimg.Color{o.Color[0], o.Color[1], o.Color[2]}
    }

    return Process(buf, opts)
}

func watermarkImage(buf []byte, o ImageOptions) (Image, error) {

    fmt.Printf("4.0 size = %d \n", len(buf))

    aitorfile, err := os.Open(o.Image)
        if err != nil {
            fmt.Fprintf(os.Stderr, "%s\n", err)
            return Image{}, NewError("Invalid watermark image.", BadRequest)
        }
    fmt.Printf("4.1 size = %s \n", o.Image)

    imageBuf, _ := ioutil.ReadAll(aitorfile)
    if len(imageBuf) == 0 {
        return Image{}, NewError("Invalid watermark image. Buffer = 0", BadRequest)
    }
    fmt.Printf("4.2 size = %d \n", len(imageBuf))

    meta, err := bimg.Metadata(buf)
    if err != nil {
        return Image{}, NewError("Cannot retrieve image metadata: %s"+err.Error(), BadRequest)
    }
    metawatermark , err := bimg.Metadata(imageBuf)
    if err != nil {
        return Image{}, NewError("Cannot retrieve image metadata: %s"+err.Error(), BadRequest)
    }

    fmt.Printf("4.3 size = %+v \n", meta)
    fmt.Printf("4.4 size = %+v \n", metawatermark)

    fmt.Printf("4.5 size = %d \n", len(buf))

    var origimagwidth = meta.Size.Width;
    var origimagheight = meta.Size.Height;
    var waterimagwidth = metawatermark.Size.Width;
    var waterimagheight = metawatermark.Size.Height;

  /* fmt.Printf("%+v\n",meta);
    fmt.Printf("%s\n",origimagwidth);
    fmt.Printf("%s\n",origimagheight);
    fmt.Printf("%s\n",waterimagwidth);
    fmt.Printf("%s\n",waterimagheight);
    */

    var settop = (origimagheight/2) - (waterimagheight/2)
    var setleft = (origimagwidth/2) - (waterimagwidth/2)

//    fmt.Printf("%s\n",settop);
//    fmt.Printf("%s\n",setleft);

//    fmt.Printf("%+v\n",o);

    opts := BimgOptions(o)
//    opts.WatermarkImage.Left = o.Left;
//    opts.WatermarkImage.Top = o.Top;
    opts.WatermarkImage.Left = setleft;
    opts.WatermarkImage.Top = settop;
    opts.WatermarkImage.Buf = imageBuf;
    opts.WatermarkImage.Opacity = o.Opacity;
    opts.WatermarkImage.Gravity = 1;

    fmt.Printf("5. size = %d \n", len(buf))

    return Process(buf, opts)
}

func GaussianBlur(buf []byte, o ImageOptions) (Image, error) {
    if o.Sigma == 0 && o.MinAmpl == 0 {
        return Image{}, NewError("Missing required param: sigma or minampl", BadRequest)
    }
    opts := BimgOptions(o)
    return Process(buf, opts)
}

func Pipeline(buf []byte, o ImageOptions) (Image, error) {
    if len(o.Operations) == 0 {
        return Image{}, NewError("Missing or invalid pipeline operations JSON", BadRequest)
    }
    if len(o.Operations) > 10 {
        return Image{}, NewError("Maximum allowed pipeline operations exceeded", BadRequest)
    }

    // Validate and built operations
    for i, operation := range o.Operations {
        // Normalize operation name
        name := strings.TrimSpace(strings.ToLower(operation.Name))

        // Validate supported operation name
        var exists bool
        if operation.Operation, exists = OperationsMap[operation.Name]; !exists {
            return Image{}, NewError(fmt.Sprintf("Unsupported operation name: %s", name), BadRequest)
        }

        // Parse and construct operation options
        operation.ImageOptions = readMapParams(operation.Params)

        // Mutate list by value
        o.Operations[i] = operation
    }

    var image Image
    var err error

    // Reduce image by running multiple operations
    image = Image{Body: buf}
    for _, operation := range o.Operations {
        var curImage Image
        curImage, err = operation.Operation(image.Body, operation.ImageOptions)
        if err != nil && !operation.IgnoreFailure {
            return Image{}, err
        }
        if operation.IgnoreFailure {
            err = nil
        }
        if err == nil {
            image = curImage
        }
    }

    return image, err
}

func Process(buf []byte, opts bimg.Options) (out Image, err error) {
    defer func() {
        if r := recover(); r != nil {
            switch value := r.(type) {
            case error:
                err = value
            case string:
                err = errors.New(value)
            default:
                err = errors.New("libvips internal error")
            }
            out = Image{}
        }
    }()
    buforig := buf

    fmt.Printf("6.0 size = %d \n", len(buf))
    fmt.Printf("6.0 opts = %+v \n", opts)

    buf, err = bimg.Resize(buforig, opts)
    if err != nil {
        fmt.Printf("Error converting the image: %s. Serving original image.\n", err);
        mime := GetImageMimeType(bimg.DetermineImageType(buf))
        return Image{Body: buforig, Mime: mime}, nil
        //return Image{}, err
    }
    fmt.Printf("6.1 size = %d \n", len(buf))

    mime := GetImageMimeType(bimg.DetermineImageType(buf))
    return Image{Body: buf, Mime: mime}, nil
}
