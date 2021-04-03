#include <node.h>
#include "sstablekeys.h"

namespace sstj {

using v8::Exception;
using v8::FunctionCallbackInfo;
using v8::Isolate;
using v8::Local;
using v8::Object;
using v8::String;
using v8::Number;
using v8::Value;

const char* ToCString(const String::Utf8Value& value) {
  return *value ? *value : "<string conversion failed>";
}

void Method(const FunctionCallbackInfo<Value>& args) {
  Isolate* isolate = args.GetIsolate();
  if (args.Length() < 2) {
    // Throw an Error that is passed back to JavaScript
    isolate->ThrowException(Exception::TypeError(
        String::NewFromUtf8(isolate,
                            "Wrong number of arguments").ToLocalChecked()));
    return;
  }
  // Check the argument types
  if (!args[0]->IsString() || !args[1]->IsNumber() || !args[2]->IsNumber()) {
    isolate->ThrowException(Exception::TypeError(
        String::NewFromUtf8(isolate,
                            "Wrong arguments").ToLocalChecked()));
    return;
  }

  v8::String::Utf8Value path(isolate, args[0]);
  const char* cpath = ToCString(path);
  char* result = SSTJ((char*)cpath, args[1].As<Number>()->Value(), args[2].As<Number>()->Value());

  //v8::String::Utf8Value v8result(isolate, result);
  //std::string v8result(result);
  args.GetReturnValue().Set(String::NewFromUtf8(isolate, result).ToLocalChecked());
}

void init(Local<Object> exports) {
  NODE_SET_METHOD(exports, "sstj", Method);
}

NODE_MODULE(NODE_GYP_MODULE_NAME, init)

}  // namespace sstj
