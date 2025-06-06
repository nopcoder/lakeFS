/*
 * lakeFS API
 * lakeFS HTTP API
 *
 * The version of the OpenAPI document: 1.0.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */


package io.lakefs.clients.sdk.model;

import java.util.Objects;
import com.google.gson.TypeAdapter;
import com.google.gson.annotations.JsonAdapter;
import com.google.gson.annotations.SerializedName;
import com.google.gson.stream.JsonReader;
import com.google.gson.stream.JsonWriter;
import java.io.IOException;
import java.util.Arrays;

import com.google.gson.Gson;
import com.google.gson.GsonBuilder;
import com.google.gson.JsonArray;
import com.google.gson.JsonDeserializationContext;
import com.google.gson.JsonDeserializer;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import com.google.gson.JsonParseException;
import com.google.gson.TypeAdapterFactory;
import com.google.gson.reflect.TypeToken;
import com.google.gson.TypeAdapter;
import com.google.gson.stream.JsonReader;
import com.google.gson.stream.JsonWriter;
import java.io.IOException;

import java.lang.reflect.Type;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

import io.lakefs.clients.sdk.JSON;

/**
 * CopyPartSource
 */
@javax.annotation.Generated(value = "org.openapitools.codegen.languages.JavaClientCodegen")
public class CopyPartSource {
  public static final String SERIALIZED_NAME_REPOSITORY = "repository";
  @SerializedName(SERIALIZED_NAME_REPOSITORY)
  private String repository;

  public static final String SERIALIZED_NAME_REF = "ref";
  @SerializedName(SERIALIZED_NAME_REF)
  private String ref;

  public static final String SERIALIZED_NAME_PATH = "path";
  @SerializedName(SERIALIZED_NAME_PATH)
  private String path;

  public static final String SERIALIZED_NAME_RANGE = "range";
  @SerializedName(SERIALIZED_NAME_RANGE)
  private String range;

  public CopyPartSource() {
  }

  public CopyPartSource repository(String repository) {
    
    this.repository = repository;
    return this;
  }

   /**
   * Get repository
   * @return repository
  **/
  @javax.annotation.Nonnull
  public String getRepository() {
    return repository;
  }


  public void setRepository(String repository) {
    this.repository = repository;
  }


  public CopyPartSource ref(String ref) {
    
    this.ref = ref;
    return this;
  }

   /**
   * Get ref
   * @return ref
  **/
  @javax.annotation.Nonnull
  public String getRef() {
    return ref;
  }


  public void setRef(String ref) {
    this.ref = ref;
  }


  public CopyPartSource path(String path) {
    
    this.path = path;
    return this;
  }

   /**
   * Get path
   * @return path
  **/
  @javax.annotation.Nonnull
  public String getPath() {
    return path;
  }


  public void setPath(String path) {
    this.path = path;
  }


  public CopyPartSource range(String range) {
    
    this.range = range;
    return this;
  }

   /**
   * Range of bytes to copy
   * @return range
  **/
  @javax.annotation.Nullable
  public String getRange() {
    return range;
  }


  public void setRange(String range) {
    this.range = range;
  }

  /**
   * A container for additional, undeclared properties.
   * This is a holder for any undeclared properties as specified with
   * the 'additionalProperties' keyword in the OAS document.
   */
  private Map<String, Object> additionalProperties;

  /**
   * Set the additional (undeclared) property with the specified name and value.
   * If the property does not already exist, create it otherwise replace it.
   *
   * @param key name of the property
   * @param value value of the property
   * @return the CopyPartSource instance itself
   */
  public CopyPartSource putAdditionalProperty(String key, Object value) {
    if (this.additionalProperties == null) {
        this.additionalProperties = new HashMap<String, Object>();
    }
    this.additionalProperties.put(key, value);
    return this;
  }

  /**
   * Return the additional (undeclared) property.
   *
   * @return a map of objects
   */
  public Map<String, Object> getAdditionalProperties() {
    return additionalProperties;
  }

  /**
   * Return the additional (undeclared) property with the specified name.
   *
   * @param key name of the property
   * @return an object
   */
  public Object getAdditionalProperty(String key) {
    if (this.additionalProperties == null) {
        return null;
    }
    return this.additionalProperties.get(key);
  }


  @Override
  public boolean equals(Object o) {
    if (this == o) {
      return true;
    }
    if (o == null || getClass() != o.getClass()) {
      return false;
    }
    CopyPartSource copyPartSource = (CopyPartSource) o;
    return Objects.equals(this.repository, copyPartSource.repository) &&
        Objects.equals(this.ref, copyPartSource.ref) &&
        Objects.equals(this.path, copyPartSource.path) &&
        Objects.equals(this.range, copyPartSource.range)&&
        Objects.equals(this.additionalProperties, copyPartSource.additionalProperties);
  }

  @Override
  public int hashCode() {
    return Objects.hash(repository, ref, path, range, additionalProperties);
  }

  @Override
  public String toString() {
    StringBuilder sb = new StringBuilder();
    sb.append("class CopyPartSource {\n");
    sb.append("    repository: ").append(toIndentedString(repository)).append("\n");
    sb.append("    ref: ").append(toIndentedString(ref)).append("\n");
    sb.append("    path: ").append(toIndentedString(path)).append("\n");
    sb.append("    range: ").append(toIndentedString(range)).append("\n");
    sb.append("    additionalProperties: ").append(toIndentedString(additionalProperties)).append("\n");
    sb.append("}");
    return sb.toString();
  }

  /**
   * Convert the given object to string with each line indented by 4 spaces
   * (except the first line).
   */
  private String toIndentedString(Object o) {
    if (o == null) {
      return "null";
    }
    return o.toString().replace("\n", "\n    ");
  }


  public static HashSet<String> openapiFields;
  public static HashSet<String> openapiRequiredFields;

  static {
    // a set of all properties/fields (JSON key names)
    openapiFields = new HashSet<String>();
    openapiFields.add("repository");
    openapiFields.add("ref");
    openapiFields.add("path");
    openapiFields.add("range");

    // a set of required properties/fields (JSON key names)
    openapiRequiredFields = new HashSet<String>();
    openapiRequiredFields.add("repository");
    openapiRequiredFields.add("ref");
    openapiRequiredFields.add("path");
  }

 /**
  * Validates the JSON Element and throws an exception if issues found
  *
  * @param jsonElement JSON Element
  * @throws IOException if the JSON Element is invalid with respect to CopyPartSource
  */
  public static void validateJsonElement(JsonElement jsonElement) throws IOException {
      if (jsonElement == null) {
        if (!CopyPartSource.openapiRequiredFields.isEmpty()) { // has required fields but JSON element is null
          throw new IllegalArgumentException(String.format("The required field(s) %s in CopyPartSource is not found in the empty JSON string", CopyPartSource.openapiRequiredFields.toString()));
        }
      }

      // check to make sure all required properties/fields are present in the JSON string
      for (String requiredField : CopyPartSource.openapiRequiredFields) {
        if (jsonElement.getAsJsonObject().get(requiredField) == null) {
          throw new IllegalArgumentException(String.format("The required field `%s` is not found in the JSON string: %s", requiredField, jsonElement.toString()));
        }
      }
        JsonObject jsonObj = jsonElement.getAsJsonObject();
      if (!jsonObj.get("repository").isJsonPrimitive()) {
        throw new IllegalArgumentException(String.format("Expected the field `repository` to be a primitive type in the JSON string but got `%s`", jsonObj.get("repository").toString()));
      }
      if (!jsonObj.get("ref").isJsonPrimitive()) {
        throw new IllegalArgumentException(String.format("Expected the field `ref` to be a primitive type in the JSON string but got `%s`", jsonObj.get("ref").toString()));
      }
      if (!jsonObj.get("path").isJsonPrimitive()) {
        throw new IllegalArgumentException(String.format("Expected the field `path` to be a primitive type in the JSON string but got `%s`", jsonObj.get("path").toString()));
      }
      if ((jsonObj.get("range") != null && !jsonObj.get("range").isJsonNull()) && !jsonObj.get("range").isJsonPrimitive()) {
        throw new IllegalArgumentException(String.format("Expected the field `range` to be a primitive type in the JSON string but got `%s`", jsonObj.get("range").toString()));
      }
  }

  public static class CustomTypeAdapterFactory implements TypeAdapterFactory {
    @SuppressWarnings("unchecked")
    @Override
    public <T> TypeAdapter<T> create(Gson gson, TypeToken<T> type) {
       if (!CopyPartSource.class.isAssignableFrom(type.getRawType())) {
         return null; // this class only serializes 'CopyPartSource' and its subtypes
       }
       final TypeAdapter<JsonElement> elementAdapter = gson.getAdapter(JsonElement.class);
       final TypeAdapter<CopyPartSource> thisAdapter
                        = gson.getDelegateAdapter(this, TypeToken.get(CopyPartSource.class));

       return (TypeAdapter<T>) new TypeAdapter<CopyPartSource>() {
           @Override
           public void write(JsonWriter out, CopyPartSource value) throws IOException {
             JsonObject obj = thisAdapter.toJsonTree(value).getAsJsonObject();
             obj.remove("additionalProperties");
             // serialize additional properties
             if (value.getAdditionalProperties() != null) {
               for (Map.Entry<String, Object> entry : value.getAdditionalProperties().entrySet()) {
                 if (entry.getValue() instanceof String)
                   obj.addProperty(entry.getKey(), (String) entry.getValue());
                 else if (entry.getValue() instanceof Number)
                   obj.addProperty(entry.getKey(), (Number) entry.getValue());
                 else if (entry.getValue() instanceof Boolean)
                   obj.addProperty(entry.getKey(), (Boolean) entry.getValue());
                 else if (entry.getValue() instanceof Character)
                   obj.addProperty(entry.getKey(), (Character) entry.getValue());
                 else {
                   obj.add(entry.getKey(), gson.toJsonTree(entry.getValue()).getAsJsonObject());
                 }
               }
             }
             elementAdapter.write(out, obj);
           }

           @Override
           public CopyPartSource read(JsonReader in) throws IOException {
             JsonElement jsonElement = elementAdapter.read(in);
             validateJsonElement(jsonElement);
             JsonObject jsonObj = jsonElement.getAsJsonObject();
             // store additional fields in the deserialized instance
             CopyPartSource instance = thisAdapter.fromJsonTree(jsonObj);
             for (Map.Entry<String, JsonElement> entry : jsonObj.entrySet()) {
               if (!openapiFields.contains(entry.getKey())) {
                 if (entry.getValue().isJsonPrimitive()) { // primitive type
                   if (entry.getValue().getAsJsonPrimitive().isString())
                     instance.putAdditionalProperty(entry.getKey(), entry.getValue().getAsString());
                   else if (entry.getValue().getAsJsonPrimitive().isNumber())
                     instance.putAdditionalProperty(entry.getKey(), entry.getValue().getAsNumber());
                   else if (entry.getValue().getAsJsonPrimitive().isBoolean())
                     instance.putAdditionalProperty(entry.getKey(), entry.getValue().getAsBoolean());
                   else
                     throw new IllegalArgumentException(String.format("The field `%s` has unknown primitive type. Value: %s", entry.getKey(), entry.getValue().toString()));
                 } else if (entry.getValue().isJsonArray()) {
                     instance.putAdditionalProperty(entry.getKey(), gson.fromJson(entry.getValue(), List.class));
                 } else { // JSON object
                     instance.putAdditionalProperty(entry.getKey(), gson.fromJson(entry.getValue(), HashMap.class));
                 }
               }
             }
             return instance;
           }

       }.nullSafe();
    }
  }

 /**
  * Create an instance of CopyPartSource given an JSON string
  *
  * @param jsonString JSON string
  * @return An instance of CopyPartSource
  * @throws IOException if the JSON string is invalid with respect to CopyPartSource
  */
  public static CopyPartSource fromJson(String jsonString) throws IOException {
    return JSON.getGson().fromJson(jsonString, CopyPartSource.class);
  }

 /**
  * Convert an instance of CopyPartSource to an JSON string
  *
  * @return JSON string
  */
  public String toJson() {
    return JSON.getGson().toJson(this);
  }
}

