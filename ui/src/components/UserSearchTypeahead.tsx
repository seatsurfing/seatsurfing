import React from "react";
import { AsyncTypeahead } from "react-bootstrap-typeahead";
import "react-bootstrap-typeahead/css/Typeahead.css";
import ProfilePicture from "@/components/ProfilePicture";
import { TranslationFunc } from "@/components/withTranslation";
import RendererUtils from "@/util/RendererUtils";
import Search, { SearchOptions } from "@/types/Search";

interface Props {
  t: TranslationFunc;
  id?: string;
  inputProps?: any;
  disabled?: boolean;
  multiple: boolean;
  selected?: any[];
  defaultSelected?: any[];
  onChange: (selected: any[]) => void;
  searchFn?: (query: string) => Promise<any[]>;
}

interface State {
  options: any[];
  loading: boolean;
}

class UserSearchTypeahead extends React.Component<Props, State> {
  private typeahead: any = null;

  constructor(props: Props) {
    super(props);
    this.state = {
      options: [],
      loading: false,
    };
  }

  filterBy = () => {
    return true;
  };

  handleSearch = (query: string) => {
    this.setState({ loading: true });
    const search = this.props.searchFn
      ? this.props.searchFn(query)
      : this.defaultSearch(query);
    search.then((options) => {
      this.setState({ options, loading: false });
    });
  };

  defaultSearch = (query: string): Promise<any[]> => {
    const options = new SearchOptions();
    options.includeUsers = true;
    options.keyword = query ? query : "";
    return Search.search(options).then((res) => res.users);
  };

  clear = () => {
    this.typeahead?.clear();
  };

  render() {
    return (
      <AsyncTypeahead
        disabled={this.props.disabled}
        filterBy={this.filterBy}
        id={this.props.id}
        inputProps={this.props.inputProps}
        isLoading={this.state.loading}
        labelKey="email"
        multiple={this.props.multiple}
        minLength={3}
        selected={this.props.selected}
        defaultSelected={this.props.defaultSelected}
        onChange={this.props.onChange}
        onSearch={this.handleSearch}
        options={this.state.options}
        placeholder={this.props.t("searchForUser")}
        ref={(ref: any) => {
          this.typeahead = ref;
        }}
        renderMenuItemChildren={(option: any) => (
          <div className="d-flex">
            <ProfilePicture width={24} height={24} />
            <span style={{ marginLeft: "10px" }}>
              {option.email}
              {RendererUtils.preAndSuffixIfDefined(
                RendererUtils.fullname(option.firstname, option.lastname),
                " (",
                ")",
              )}{" "}
            </span>
          </div>
        )}
      />
    );
  }
}

export default UserSearchTypeahead;
