package http

import (
	"net/http"
)

const MethodGet = http.MethodGet
const StatusContinue = http.StatusContinue
const DefaultMaxHeaderBytes = http.DefaultMaxHeaderBytes
const DefaultMaxIdleConnsPerHost = http.DefaultMaxIdleConnsPerHost
const TimeFormat = http.TimeFormat
const TrailerPrefix = http.TrailerPrefix

var ErrNotSupported = http.ErrNotSupported
var ErrBodyNotAllowed = http.ErrBodyNotAllowed
var ServerContextKey = http.ServerContextKey
var ErrAbortHandler = http.ErrAbortHandler
var ErrBodyReadAfterClose = http.ErrBodyReadAfterClose
var ErrHandlerTimeout = http.ErrHandlerTimeout
var ErrLineTooLong = http.ErrLineTooLong
var ErrMissingFile = http.ErrMissingFile
var ErrNoCookie = http.ErrNoCookie
var ErrNoLocation = http.ErrNoLocation
var ErrSchemeMismatch = http.ErrSchemeMismatch
var ErrServerClosed = http.ErrServerClosed
var ErrSkipAltProtocol = http.ErrSkipAltProtocol
var ErrUseLastResponse = http.ErrUseLastResponse
var NoBody = http.NoBody
var CanonicalHeaderKey = http.CanonicalHeaderKey
var DetectContentType = http.DetectContentType
var Error = http.Error
var Get = http.Get
var Head = http.Head
var ListenAndServe = http.ListenAndServe
var ListenAndServeTLS = http.ListenAndServeTLS
var MaxBytesReader = http.MaxBytesReader
var NewRequest = http.NewRequest
var NewRequestWithContext = http.NewRequestWithContext
var NotFound = http.NotFound
var ParseCookie = http.ParseCookie
var ParseHTTPVersion = http.ParseHTTPVersion
var ParseSetCookie = http.ParseSetCookie
var ParseTime = http.ParseTime
var Post = http.Post
var PostForm = http.PostForm
var ProxyFromEnvironment = http.ProxyFromEnvironment
var ProxyURL = http.ProxyURL
var ReadRequest = http.ReadRequest
var ReadResponse = http.ReadResponse
var Redirect = http.Redirect
var Serve = http.Serve
var ServeContent = http.ServeContent
var ServeFile = http.ServeFile
var ServeFileFS = http.ServeFileFS
var ServeTLS = http.ServeTLS
var SetCookie = http.SetCookie
var StatusText = http.StatusText

type Client = http.Client
type CloseNotifier = http.CloseNotifier
type ConnState = http.ConnState

const StateNew = http.StateNew

type Cookie = http.Cookie
type CookieJar = http.CookieJar
type Dir = http.Dir
type Fil = http.File
type FileSystem = http.FileSystem

var FS = http.FS

type Flusher = http.Flusher
type Handler = http.Handler

var AllowQuerySemicolons = http.AllowQuerySemicolons
var FileServer = http.FileServer
var FileServerFS = http.FileServerFS
var MaxBytesHandler = http.MaxBytesHandler
var NotFoundHandler = http.NotFoundHandler
var RedirectHandler = http.RedirectHandler
var StripPrefix = http.StripPrefix
var TimeoutHandler = http.TimeoutHandler

type Header = http.Header
type Hijacker = http.Hijacker
type MaxBytesError = http.MaxBytesError
type ProtocolError = http.ProtocolError
type PushOptions = http.PushOptions
type Pusher = http.Pusher
type Response = http.Response
type ResponseController = http.ResponseController

var NewResponseController = http.NewResponseController

type ResponseWriter = http.ResponseWriter
type RoundTripper = http.RoundTripper

var DefaultTransport = http.DefaultTransport
var NewFileTransport = http.NewFileTransport
var NewFileTransportFS = http.NewFileTransportFS

type SameSite = http.SameSite

const SameSiteDefaultMode = http.SameSiteDefaultMode

type Server = http.Server
type Transport = http.Transport
